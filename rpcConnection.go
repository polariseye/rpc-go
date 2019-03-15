package rpc

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

type RpcConnection struct {
	container      *RpcContainer
	frameContainer FrameContainer
	con            net.Conn
	isClosed       bool
	sendChan       chan *DataFrame

	// 数据的字节序
	byteOrder binary.ByteOrder

	requestId uint32
}

func (this *RpcConnection) Call(methodName string, requestMsg proto.Message, responseMsg proto.Message) (err error) {
	err, downChan := this.Go(methodName, requestMsg, responseMsg)
	if err != nil {
		return err
	}

	return <-downChan
}

func (this *RpcConnection) Go(methodName string, requestMsg proto.Message, responseMsg proto.Message) (err error, donChan <-chan error) {
	var requestBytes []byte
	if requestMsg != nil {
		requestBytes, err = proto.Marshal(requestMsg)
	}

	// 添加到等待应答的列表中
	requestObj := &RequestInfo{
		RequestId: this.getRequestId(),
		DownChan:  make(chan error, 10),
		ReturnObj: responseMsg,
	}
	frameObj := newRequestFrame(requestObj, methodName, requestBytes, requestObj.RequestId, true)

	this.frameContainer.AddRequest(requestObj)
	this.sendChan <- frameObj

	return nil, requestObj.DownChan
}

func (this *RpcConnection) getRequestId() uint32 {
	return atomic.AddUint32(&this.requestId, 1)
}

func (this *RpcConnection) Close() {

}

func (this *RpcConnection) receive() {
	var header = make([]byte, HEADER_LENGTH)
	for this.isClosed == false {
		// 读取包头
		err := this.receiveHeader(this.con, header)
		if err != nil {
			this.isClosed = true
			break
		}

		// 获取帧头
		frameObj := convertHeader(header, this.byteOrder)
		// 是请求帧，但又没有设置请求函数，则代表是非法帧
		if frameObj.MethodNameLen == 0 && frameObj.ResponseFrameId == 0 {
			// 跳过错误的帧
			continue
		}

		// 读取包内容
		if frameObj.MethodNameLen > 0 || frameObj.ContentLength > 0 {
			buffer := make([]byte, frameObj.ContentLength+uint32(frameObj.MethodNameLen))
			_, err = io.ReadFull(this.con, buffer)
			if err != nil {
				this.isClosed = true
				break
			}
		}

		// 处理请求
		this.HandleFrame(frameObj)
	}
}

func (this *RpcConnection) receiveHeader(con net.Conn, header []byte) error {
	startIndex := 1
	for this.isClosed == false {
		_, err := io.ReadFull(this.con, header[:1])
		if err != nil {
			return err
		}
		if header[0] != HEADER { //// 找协议头
			continue
		}

		for this.isClosed == false {
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
	for this.isClosed == false {
		select {
		case item := <-this.sendChan:
			_, err := this.con.Write(item.GetHeader(this.byteOrder))
			if err != nil {
				item.RequestObj.ErrObj = err
				item.RequestObj.DownChan <- err
				this.isClosed = true
				break
			}

			if item.MethodNameLen > 0 {
				_, err = this.con.Write(item.MethodNameBytes)
				if err != nil {
					// 当前包的异常处理
					item.RequestObj.ErrObj = err
					item.RequestObj.DownChan <- err
					this.isClosed = true
					break
				}
			}

			if item.ContentLength > 0 {
				_, err = this.con.Write(item.Data)
				if err != nil {
					// 当前包的异常处理
					item.RequestObj.ErrObj = err
					item.RequestObj.DownChan <- err
					this.isClosed = true
					break
				}
			}
		}
	}
}

func (this *RpcConnection) HandleFrame(frameObj *DataFrame) {
	if frameObj.ResponseFrameId != 0 {
		// 应答处理
		requestObj, exist := this.frameContainer.GetRequestInfo(frameObj.ResponseFrameId)
		if exist == false {
			// 丢掉
			return
		}

		requestObj.ReturnBytes = frameObj.Data
		if frameObj.IsError() {
			requestObj.ErrObj = errors.New(string(frameObj.Data))
		} else if requestObj.ReturnObj != nil {
			requestObj.ErrObj = proto.Unmarshal(frameObj.Data, requestObj.ReturnObj.(proto.Message))
		}

		// 通知完成
		requestObj.DownChan <- requestObj.ErrObj
	} else {
		// 请求处理
		methodObj, exist := this.container.getMethod(frameObj.MethodName())
		if exist == false {
			return
		}

		returnBytes, err := methodObj.Invoke(frameObj.Data, this.byteOrder)
		if err != nil {

		}
	}

	return
}

func NewRpcConnection(container *RpcContainer, con net.Conn) *RpcConnection {
	var result = &RpcConnection{
		container: container,
		con:       con,
		requestId: rand.New(rand.NewSource(time.Now().Unix())).Uint32(), //// 产生一个随机数
	}

	// 开协程进行具体处理
	go result.receive()
	go result.send()

	return result
}
