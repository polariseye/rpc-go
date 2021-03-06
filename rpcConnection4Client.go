package rpc

import (
	"reflect"
	"time"

	"github.com/polariseye/rpc-go/log"
)

type RpcConnection4Client struct {
	*RpcConnection
	*RpcWatchBase

	keepAliveInterval    int64 //// 单位：秒 默认5秒
	preSendKeepAliveTime int64

	connectionedHandlerData map[string]func(connObj RpcConnectioner)
}

func (this *RpcConnection4Client) afterSend(frameObj *DataFrame) (err error) {
	this.invokeAfterSendHandler(this, frameObj)
	return nil
}

func (this *RpcConnection4Client) sendSchedule() (err error) {
	now := time.Now().Unix()
	connObj := this.RpcConnection

	// 心跳发送
	if connObj != nil && (now-this.preSendKeepAliveTime) > this.keepAliveInterval {
		frameObj := newRequestFrame(nil, "", nil, connObj.getRequestId(), true)
		frameObj.SetTransformType(TransformType_KeepAlive)

		this.directlySendFrame(frameObj)
		//// 此处不管是否报错，都需要加心跳，以避免一直发不停心跳
		this.preSendKeepAliveTime = now
	}

	this.invokeSendScheduleHandler(this)

	return
}

func (this *RpcConnection4Client) beforeHandleFrame(frameObj *DataFrame) (isHandled bool, err error) {
	if frameObj.TransformType() == TransformType_KeepAlive {
		log.Debug("receive keepalive Addr:%v", this.Addr())
		// 心跳不下发了，没什么意思
		return
	}

	return this.invokeBeforeHandleFrameHandler(this, frameObj)
}

func (this *RpcConnection4Client) afterInvoke(frameObj *DataFrame, returnList []reflect.Value, err error) (resultReturnList []reflect.Value, resultErr error) {
	return this.invokeAfterInvokeHandler(this, frameObj, returnList, err)
}

func (this *RpcConnection4Client) afterClose() {
	this.invokeCloseHandler(this)
}
func (this *RpcConnection4Client) setConnection(con *RpcConnection) {
	this.RpcConnection = con

	// 触发连接事件
	this.invokeConnectedHandler(con)
}

func (this *RpcConnection4Client) IsClosed() bool {
	if this.RpcConnection == nil {
		return true
	}

	return this.RpcConnection.IsClosed()
}

func (this *RpcConnection4Client) AddConnectedHandler(funcName string, funcObj func(connObj RpcConnectioner)) (err error) {
	if _, exist := this.connectionedHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.connectionedHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcConnection4Client) invokeConnectedHandler(connObj RpcConnectioner) {
	for _, item := range this.connectionedHandlerData {
		item(connObj)
	}
}

func NewRpcConnection4Client() *RpcConnection4Client {
	result := &RpcConnection4Client{
		RpcWatchBase:            newRpcWatchBase(),
		keepAliveInterval:       5,
		connectionedHandlerData: make(map[string]func(connObj RpcConnectioner), 4),
	}

	return result
}
