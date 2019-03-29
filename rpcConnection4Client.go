package rpc

import (
	"time"
)

type RpcConnection4Client struct {
	*RpcConnection
	*RpcWatchBase

	keepAliveInterval    int64 //// 单位：秒 默认5秒
	preSendKeepAliveTime int64
}

func (this *RpcConnection4Client) afterSend(frameObj *DataFrame) (err error) {
	this.invokeAfterSendHandler(this, frameObj)
	return nil
}

func (this *RpcConnection4Client) sendSchedule() (err error) {
	now := time.Now().Unix()

	// 心跳发送
	if (now - this.preSendKeepAliveTime) > this.keepAliveInterval {
		frameObj := newRequestFrame(nil, "", nil, this.getRequestId(), true)
		frameObj.SetTransformType(TransformType_KeepAlive)
		_, err := this.con.Write(frameObj.GetHeader(this.byteOrder))
		if err != nil {
			return err
		}

		this.preSendKeepAliveTime = now
	}

	this.invokeSendScheduleHandler(this)

	return
}

func (this *RpcConnection4Client) beforeHandleFrame(frameObj *DataFrame) (isHandled bool, err error) {
	if frameObj.TransformType() == TransformType_KeepAlive {
		// 心跳不下发了，没什么意思
		return
	}

	this.invokeBeforeHandleFrameHandler(this, frameObj)
	return false, nil
}

func (this *RpcConnection4Client) afterInvoke(frameObj *DataFrame, returnBytes []byte, err error) {
	this.invokeAfterInvokeHandler(this, returnBytes, err)
}

func (this *RpcConnection4Client) afterClose() {
	this.invokeCloseHandler(this)
}
func (this *RpcConnection4Client) setConnection(con *RpcConnection) {
	this.RpcConnection = con
}

func (this *RpcConnection4Client) IsClosed() bool {
	if this.RpcConnection == nil {
		return false
	}

	return this.RpcConnection.IsClosed()
}

func NewRpcConnection4Client() *RpcConnection4Client {
	result := &RpcConnection4Client{
		RpcWatchBase: newRpcWatchBase(),
	}

	return result
}
