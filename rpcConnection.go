package rpc

import (
	"errors"
	"io"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/polariseye/rpc-go/log"
)

type RpcConnection struct {
	container      *RpcContainer
	frameContainer *FrameContainer
	con            net.Conn
	isClosed       int32
	sendChan       chan *DataFrame
	requestChan    chan *DataFrame

	requestId uint32
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
		requestBytes, err = this.container.getConvertorFunc().MarshalValue(requestObj...)
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
		ExpireTime: time.Now().UnixNano()/1000000 + this.container.requestExpireMillisecond,
	}
	frameObj := newRequestFrame(requestInfoObj, methodName, requestBytes, requestInfoObj.RequestId, true)

	this.frameContainer.AddRequest(requestInfoObj)
	this.sendChan <- frameObj

	return requestInfoObj.DownChan, nil
}

func (this *RpcConnection) CallAsyncWithNoResponse(methodName string, requestObj []interface{}, responseObj []interface{}) (err error) {
	var requestBytes []byte
	if len(requestObj) > 0 {
		requestBytes, err = this.container.getConvertorFunc().MarshalValue(requestObj...)
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
		ExpireTime: time.Now().UnixNano()/1000000 + this.container.requestExpireMillisecond,
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
		requestBytes, err = this.container.getConvertorFunc().MarshalValue(requestObj...)
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

func (this *RpcConnection) Close(err error) {
	// 设置为已关闭
	if atomic.CompareAndSwapInt32(&this.isClosed, No, Yes) == false {
		return
	}

	// 避免请求处理协程卡住，发一个nil让它流转下
	this.requestChan <- nil

	// 清空所有请求
	this.frameContainer.ReturnAllRequest(err)

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
	defer this.Close(err)

	var header = make([]byte, HEADER_LENGTH)
	for this.isClosed == No {
		// 读取包头
		err = this.receiveHeader(this.con, header)
		if err != nil {
			break
		}

		// 获取帧头
		frameObj := convertHeader(header, this.container.byteOrder)
		// 是请求帧，但又没有设置请求函数，则代表是非法帧
		if frameObj.MethodNameLen == 0 && frameObj.ResponseFrameId == 0 {
			log.Warn("receive error frame ip:%v", this.Addr())
			// 跳过错误的帧
			continue
		}

		// 读取包内容
		if frameObj.MethodNameLen > 0 || frameObj.ContentLength > 0 {
			buffer := make([]byte, frameObj.ContentLength+uint32(frameObj.MethodNameLen))
			_, err = io.ReadFull(this.con, buffer)
			if err != nil {
				break
			}

			frameObj.SetData(buffer)
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
	// 清空所有请求
	defer func() {
		for {
			select {
			case item := <-this.sendChan:
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
	defer this.Close(err)

	for this.isClosed == No {
		select {
		case item := <-this.sendChan:
			_, err = this.con.Write(item.GetHeader(this.container.byteOrder))
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
		default:
			{
				time.Sleep(5 * time.Millisecond)
			}
		}

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
			tmpErr := this.container.getConvertorFunc().UnMarhsalValue(frameObj.Data, requestObj.ReturnObj...)
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
	for this.isClosed == No {
		select {
		case frameObj := <-this.requestChan:
			{
				if frameObj == nil {
					continue
				}

				// 请求处理
				methodObj, exist := this.container.getMethod(frameObj.MethodName())
				if exist == false {
					this.response(frameObj, nil, MethodNotFoundError)
					log.Error("not fount method methodname:%v", frameObj.MethodName())

					continue
				}

				returnBytes, err := methodObj.Invoke(this, this.container.getConvertorFunc(), frameObj.Data, this.container.byteOrder, frameObj.IsNeedResponse())
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

func NewRpcConnection(container *RpcContainer, con net.Conn) *RpcConnection {
	var result = &RpcConnection{
		container:      container,
		frameContainer: newFrameContainer(),
		con:            con,
		isClosed:       No,
		sendChan:       make(chan *DataFrame, 1024),
		requestChan:    make(chan *DataFrame, 1024),
		requestId:      rand.New(rand.NewSource(time.Now().Unix())).Uint32(), //// 产生一个随机数
	}

	// 开协程进行具体处理
	go result.receive()
	go result.send()
	go result.handleRequestFrame()

	return result
}
