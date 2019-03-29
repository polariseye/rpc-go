package rpc

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/polariseye/rpc-go/log"
)

type RpcConnectioner interface {
	Call(methodName string, requestObj []interface{}, responseObj []interface{}) (err error)
	CallAsync(methodName string, requestObj []interface{}, responseObj []interface{}) (donChan <-chan error, err error)
	CallAsyncWithNoResponse(methodName string, requestObj []interface{}, responseObj []interface{}) (err error)
	CallTimeout(methodName string, requestObj []interface{}, responseObj []interface{}, expireMillisecond int64) (err error)
	CallAsyncTimeout(methodName string, requestObj []interface{}, responseObj []interface{}, expireMillisecond int64) (donChan <-chan error, err error)
	SetRequestExpireMillisecond(requestExpireMillisecond int64)
	Close()
	Conn() net.Conn
	Addr() string
	IsClosed() bool
}

// 连接Id，用于为每个连接分配一个唯一Id
var nowMaxConnectionId int64

// 获取一个唯一连接Id
func getNextConnectionId() int64 {
	return atomic.AddInt64(&nowMaxConnectionId, 1)
}

type RpcConnection struct {
	connectionId     int64                 //// 连接Id
	apiMgr           *ApiMgr               //// Api管理对象
	frameContainer   *FrameContainer       //// 帧容器
	con              net.Conn              //// 实际连接对象
	isClosed         int32                 //// 当前是否是已经关闭了连接
	sendChan         chan *DataFrame       //// 帧发送队列
	requestChan      chan *DataFrame       //// 请求发送帧
	rpcWatcherObj    RpcWatcher            //// 连接具体处理
	byteOrder        binary.ByteOrder      //// 数据的字节序
	getConvertorFunc func() IByteConvertor //// 数据转换对象获取

	requestExpireMillisecond int64  // 请求超时时间,单位毫秒
	requestId                uint32 //// 请求Id，会为每次请求分配一个唯一Id

	closeWaitGroup sync.WaitGroup
}

// SetRequestExpireMillisecond 设置默认的请求超时时间,
// requestExpireMillisecond:请求超时时长 单位：毫秒
func (this *RpcConnection) SetRequestExpireMillisecond(requestExpireMillisecond int64) {
	this.requestExpireMillisecond = requestExpireMillisecond
}

func (this *RpcConnection) Call(methodName string, requestObj []interface{}, responseObj []interface{}) (err error) {
	downChan, err := this.CallAsync(methodName, requestObj, responseObj)
	if err != nil {
		return err
	}

	if this.isClosed != No {
		return io.EOF
	}

	return <-downChan
}

func (this *RpcConnection) CallAsync(methodName string, requestObj []interface{}, responseObj []interface{}) (donChan <-chan error, err error) {
	var requestBytes []byte
	if len(requestObj) > 0 {
		requestBytes, err = this.getConvertorFunc().MarshalValue(requestObj...)
		if err != nil {
			return nil, err
		}
	}

	if this.isClosed != No {
		return nil, io.EOF
	}

	// 添加到等待应答的列表中
	requestInfoObj := &RequestInfo{
		RequestId:  this.getRequestId(),
		DownChan:   make(chan error, 10),
		ReturnObj:  responseObj,
		ExpireTime: time.Now().UnixNano()/1000000 + this.requestExpireMillisecond,
	}
	frameObj := newRequestFrame(requestInfoObj, methodName, requestBytes, requestInfoObj.RequestId, true)

	this.frameContainer.AddRequest(requestInfoObj)
	this.sendChan <- frameObj

	return requestInfoObj.DownChan, nil
}

func (this *RpcConnection) CallAsyncWithNoResponse(methodName string, requestObj []interface{}, responseObj []interface{}) (err error) {
	var requestBytes []byte
	if len(requestObj) > 0 {
		requestBytes, err = this.getConvertorFunc().MarshalValue(requestObj...)
		if err != nil {
			return err
		}
	}

	if this.isClosed != No {
		return io.EOF
	}

	// 添加到等待应答的列表中
	requestInfoObj := &RequestInfo{
		RequestId:  this.getRequestId(),
		DownChan:   make(chan error, 10),
		ReturnObj:  responseObj,
		ExpireTime: time.Now().UnixNano()/1000000 + this.requestExpireMillisecond,
	}
	frameObj := newRequestFrame(requestInfoObj, methodName, requestBytes, requestInfoObj.RequestId, false)
	this.sendChan <- frameObj

	return nil
}

func (this *RpcConnection) CallTimeout(methodName string, requestObj []interface{}, responseObj []interface{}, expireMillisecond int64) (err error) {
	downChan, err := this.CallAsyncTimeout(methodName, requestObj, responseObj, expireMillisecond)
	if err != nil {
		return err
	}

	if this.isClosed != No {
		return io.EOF
	}

	return <-downChan
}

func (this *RpcConnection) CallAsyncTimeout(methodName string, requestObj []interface{}, responseObj []interface{}, expireMillisecond int64) (donChan <-chan error, err error) {
	var requestBytes []byte
	if len(requestObj) > 0 {
		requestBytes, err = this.getConvertorFunc().MarshalValue(requestObj...)
		if err != nil {
			return nil, err
		}
	}

	if this.isClosed != No {
		return nil, io.EOF
	}

	// 添加到等待应答的列表中
	requestInfoObj := &RequestInfo{
		RequestId:  this.getRequestId(),
		DownChan:   make(chan error, 10),
		ReturnObj:  responseObj,
		ExpireTime: time.Now().UnixNano()/1000000 + expireMillisecond,
	}
	frameObj := newRequestFrame(requestInfoObj, methodName, requestBytes, requestInfoObj.RequestId, true)

	this.frameContainer.AddRequest(requestInfoObj)
	this.sendChan <- frameObj

	return requestInfoObj.DownChan, nil
}

func (this *RpcConnection) getRequestId() uint32 {
	return atomic.AddUint32(&this.requestId, 1)
}

func (this *RpcConnection) Close() {
	this.close(CustCloseConnectionError)

	// 等待所有协程退出
	this.closeWaitGroup.Wait()
}

func (this *RpcConnection) close(err error) {
	// 设置为已关闭
	if atomic.CompareAndSwapInt32(&this.isClosed, No, Yes) == false {
		return
	}

	// 避免请求处理协程卡住，发一个nil让它流转下
	this.requestChan <- nil

	// 清空所有请求
	this.frameContainer.ReturnAllRequest(err)

	con := this.con
	if con != nil {
		con.Close()
	}

	// 连接关闭时触发的对应事件
	this.rpcWatcherObj.afterClose()

	log.Debug("connection closed ip:%v", this.Addr())
}

func (this *RpcConnection) Addr() string {
	if this.con == nil {
		return ""
	}

	return this.con.RemoteAddr().String()
}

func (this *RpcConnection) receive() {
	var err error
	defer this.closeWaitGroup.Done()
	defer this.close(err)

	var header = make([]byte, HEADER_LENGTH)
	var isHandled bool
	for this.isClosed == No {
		// 读取包头
		err = this.receiveHeader(this.con, header)
		if this.isClosed == Yes || err != nil {
			break
		}

		// 获取帧头
		frameObj := convertHeader(header, this.byteOrder)
		//// 读取包内容
		if frameObj.MethodNameLen > 0 || frameObj.ContentLength > 0 {
			buffer := make([]byte, frameObj.ContentLength+uint32(frameObj.MethodNameLen))
			_, err = io.ReadFull(this.con, buffer)
			if err != nil {
				break
			}

			frameObj.SetData(buffer)
		}

		isHandled, err = this.rpcWatcherObj.beforeHandleFrame(frameObj)
		if isHandled || err != nil {
			// 已处理，或者出现error，则跳过这个包
			continue
		}

		// 是请求帧，但又没有设置请求函数，则代表是非法帧
		if frameObj.MethodNameLen == 0 && frameObj.ResponseFrameId == 0 {
			log.Warn("receive error frame ip:%v", this.Addr())
			// 跳过错误的帧
			continue
		}

		// 处理请求
		this.handleFrame(frameObj)
	}
}

func (this *RpcConnection) receiveHeader(con net.Conn, header []byte) error {
	startIndex := 1
	for this.isClosed == No {
		_, err := io.ReadFull(this.con, header[:1])
		if err != nil {
			return err
		}
		if header[0] != HEADER { //// 找协议头
			continue
		}

		for this.isClosed == No {
			_, err = io.ReadFull(this.con, header[startIndex:])
			if err != nil {
				return err
			}
			if header[HEADER_LENGTH-1] == TAIL {
				// 已解析出了帧头
				return nil
			}

			// 在解析到不正确包的情况下，先找到一个合适的包头
			for i := 1; i < HEADER_LENGTH; i++ {
				if header[i] == HEADER {
					tmpHeader := make([]byte, HEADER_LENGTH-i)
					copy(tmpHeader, header[i:])
					copy(header, tmpHeader[:HEADER_LENGTH-i])
					startIndex = i + 1
					break
				}
			}
		}
	}

	return nil
}

func (this *RpcConnection) send() {
	defer this.closeWaitGroup.Done()

	// 清空所有请求
	defer func() {
		for {
			select {
			case item := <-this.sendChan:
				if item.RequestObj == nil {
					// 因为心跳没有请求数据，所以此处需要排队
					continue
				}

				item.RequestObj.ReturnError(io.EOF)
				this.frameContainer.RemoveRequestObj(item.RequestObj.RequestId)
			default:
				{
					return
				}
			}
		}
	}()
	var err error
	defer this.close(err)

	for this.isClosed == No {
		select {
		case item := <-this.sendChan:
			_, err = this.con.Write(item.GetHeader(this.byteOrder))
			if err != nil {
				break
			}

			if item.MethodNameLen > 0 {
				_, err = this.con.Write(item.MethodNameBytes)
				if err != nil {
					// 当前包的异常处理
					break
				}
			}

			if item.ContentLength > 0 {
				_, err = this.con.Write(item.Data)
				if err != nil {
					// 当前包的异常处理
					break
				}
			}

			// 每次发送数据后调用的接口
			this.rpcWatcherObj.afterSend(item)
		default:
			{
				time.Sleep(5 * time.Millisecond)
			}
		}

		// 发送调度处理
		if err = this.rpcWatcherObj.sendSchedule(); err != nil {
			break
		}

		// 清理过期包
		this.frameContainer.ClearExpireNode()
	}
}

func (this *RpcConnection) handleFrame(frameObj *DataFrame) {
	if frameObj.ResponseFrameId != 0 {
		// 应答处理
		requestObj, exist := this.frameContainer.GetRequestInfo(frameObj.ResponseFrameId)
		if exist == false {
			// 丢掉
			return
		}

		requestObj.ReturnBytes = frameObj.Data
		if frameObj.IsError() {
			requestObj.ReturnError(errors.New(string(frameObj.Data)))
		} else if len(requestObj.ReturnObj) > 0 {
			//// 反序列化参数
			tmpErr := this.getConvertorFunc().UnMarhsalValue(frameObj.Data, requestObj.ReturnObj...)
			requestObj.Return(requestObj.ReturnObj, frameObj.Data, tmpErr)
		} else {
			//// 没有返回值
			requestObj.Return(nil, nil, nil)
		}
	} else {
		// 使用异步方式来处理请求
		this.requestChan <- frameObj
	}

	return
}

func (this *RpcConnection) handleRequestFrame() {
	defer this.closeWaitGroup.Done()

	for this.isClosed == No {
		select {
		case frameObj := <-this.requestChan:
			{
				if frameObj == nil {
					continue
				}

				// 请求处理
				methodObj, exist := this.apiMgr.getMethod(frameObj.MethodName())
				if exist == false {
					this.response(frameObj, nil, MethodNotFoundError)
					log.Error("not fount method methodname:%v", frameObj.MethodName())

					continue
				}

				returnBytes, err := methodObj.Invoke(this, this.getConvertorFunc(), frameObj.Data, this.byteOrder, frameObj.IsNeedResponse())
				this.rpcWatcherObj.afterInvoke(frameObj, returnBytes, err)
				this.response(frameObj, returnBytes, err)
			}
		}
	}
}

func (this *RpcConnection) response(frameObj *DataFrame, returnBytes []byte, err error) {
	if frameObj.IsNeedResponse() == false {
		// 不需要应答则不处理
		return
	}

	// 应答
	responseFrame := newResponseFrame(frameObj, returnBytes, this.getRequestId())
	if err != nil {
		// 应答错误处理
		responseFrame.SetError(err.Error())
	}
	this.sendChan <- responseFrame
}

func (this *RpcConnection) Conn() net.Conn {
	return this.con
}

func (this *RpcConnection) IsClosed() bool {
	return this.isClosed == Yes
}

func (this *RpcConnection) ConnectionId() int64 {
	return this.connectionId
}

func newRpcConnection(apiMgr *ApiMgr, con net.Conn, watcherObj RpcWatcher, getConvertorFunc func() IByteConvertor) *RpcConnection {
	var result = &RpcConnection{
		apiMgr:                   apiMgr,
		frameContainer:           newFrameContainer(),
		con:                      con,
		isClosed:                 No,
		sendChan:                 make(chan *DataFrame, 1024),
		requestChan:              make(chan *DataFrame, 1024),
		requestExpireMillisecond: 2 * 60 * 1000,
		rpcWatcherObj:            watcherObj,
		requestId:                rand.New(rand.NewSource(time.Now().Unix())).Uint32(), //// 产生一个随机数
		connectionId:             getNextConnectionId(),
		byteOrder:                binary.BigEndian,
		getConvertorFunc:         getConvertorFunc,
	}

	result.closeWaitGroup.Add(3)

	// 开协程进行具体处理
	go result.receive()
	go result.send()
	go result.handleRequestFrame()

	return result
}
