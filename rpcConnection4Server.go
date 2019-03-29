package rpc

import (
	"net"
	"time"

	"github.com/polariseye/rpc-go/log"
)

type RpcConnection4Server struct {
	*RpcConnection
	*RpcWatchBase

	// 心跳超时时间：单位：秒 默认20秒
	connectionTimeoutSecond int64
	preReceiveKeepAliveTime int64 //// 上次收到心跳的时间
}

func (this *RpcConnection4Server) afterSend(frameObj *DataFrame) (err error) {
	this.invokeAfterSendHandler(this, frameObj)
	return nil
}

func (this *RpcConnection4Server) sendSchedule() (err error) {
	now := time.Now().Unix()

	// 检查心跳时间
	if (now - this.preReceiveKeepAliveTime) > this.connectionTimeoutSecond {
		// 心跳超时处理
		log.Debug("Connection Timeout IP:%v", this.Addr())
		this.close(ConnectionTimeOut)

		return
	}

	this.invokeSendScheduleHandler(this)

	return nil
}

func (this *RpcConnection4Server) beforeHandleFrame(frameObj *DataFrame) (isHandled bool, err error) {
	//// 心跳处理
	if frameObj.TransformType() == TransformType_KeepAlive {
		if frameObj.ResponseFrameId == 0 {
			// log.Debug("receive KeepAlive IP:%v", this.Addr())
			//// 只有心跳请求才返回心跳应答
			this.responseKeepAlive(frameObj)
		} else {
			//log.Debug("receive KeepAlive Response IP:%v", this.Addr())
		}
		// 更新上次心跳时间
		this.preReceiveKeepAliveTime = time.Now().Unix()

		isHandled = true

		// 心跳不下发了，没什么意思
		return
	}

	this.invokeBeforeHandleFrameHandler(this, frameObj)

	return
}

func (this *RpcConnection4Server) responseKeepAlive(frameObj *DataFrame) {
	responseFrame := newResponseFrame(frameObj, nil, this.getRequestId())
	this.sendChan <- responseFrame
}

func (this *RpcConnection4Server) afterInvoke(frameObj *DataFrame, returnBytes []byte, err error) {
	this.invokeAfterInvokeHandler(this, returnBytes, err)
}

func (this *RpcConnection4Server) afterClose() {
	this.invokeCloseHandler(this)
}

func NewRpcConnection4Server(con net.Conn, apiMgr *ApiMgr, getConvertorFunc func() IByteConvertor) *RpcConnection4Server {
	result := &RpcConnection4Server{
		RpcWatchBase:            newRpcWatchBase(),
		connectionTimeoutSecond: 20,
		preReceiveKeepAliveTime: time.Now().Unix(),
	}

	result.RpcConnection = newRpcConnection(apiMgr, con, result, getConvertorFunc)

	return result
}
