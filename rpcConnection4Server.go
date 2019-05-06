package rpc

import (
	"encoding/binary"
	"net"
	"reflect"
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

// SetConnectionTimeoutSecond 设置连接超时时间（多久没有收到心跳就断开连接）
func (this *RpcConnection4Server) SetConnectionTimeoutSecond(connectionTimeoutSecond int64) {
	this.connectionTimeoutSecond = connectionTimeoutSecond
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

	return this.invokeBeforeHandleFrameHandler(this, frameObj)
}

func (this *RpcConnection4Server) responseKeepAlive(frameObj *DataFrame) {
	// 必需先获取，再判断，因为断连时，会把连接置空
	connObj := this.RpcConnection
	if connObj == nil {
		log.Info("connection is null,can not response keepalive")
		return
	}

	connObj.sendFrame(newResponseFrame(frameObj, nil, connObj.getRequestId()))
}

func (this *RpcConnection4Server) afterInvoke(frameObj *DataFrame, returnList []reflect.Value, err error) (resultReturnList []reflect.Value, resultErr error) {
	return this.invokeAfterInvokeHandler(this, frameObj, returnList, err)
}

func (this *RpcConnection4Server) afterClose() {
	this.invokeCloseHandler(this)
}

func NewRpcConnection4Server(con net.Conn, apiMgr *ApiMgr, order binary.ByteOrder, getConvertorFunc func() IByteConvertor) *RpcConnection4Server {
	result := &RpcConnection4Server{
		RpcWatchBase:            newRpcWatchBase(),
		connectionTimeoutSecond: 20,
		preReceiveKeepAliveTime: time.Now().Unix(),
	}

	result.RpcConnection = newRpcConnection(apiMgr, con, result, result, order, getConvertorFunc)

	return result
}
