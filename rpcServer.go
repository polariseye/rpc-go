package rpc

import (
	"net"
	"sync"

	"github.com/polariseye/rpc-go/log"
)

type RpcServer struct {
	*ApiMgr
	*RpcWatchBase

	connData         map[int64]*RpcConnection4Server
	connDataLockObj  sync.RWMutex
	getConvertorFunc func() IByteConvertor

	newConnectionHandlerData map[string]func(connObj RpcConnectioner) error
}

func (this *RpcServer) GetConnection(connectionId int64) (result *RpcConnection4Server, exist bool) {
	this.connDataLockObj.RLock()
	defer this.connDataLockObj.RUnlock()

	result, exist = this.connData[connectionId]
	return
}

func (this *RpcServer) invokeNewConnectionHandler(connObj *RpcConnection4Server) {
	// 进行事件关联
	connObj.AddCloseHandler("RpcServer.CloseHandler", func(connObj RpcConnectioner) {
		// 添加到连接集合中
		this.onConnectionClose(connObj.(*RpcConnection4Server))

		this.invokeCloseHandler(connObj)
	})
	connObj.AddAfterSendHandler("RpcServer.AfterSendHandler", func(connObj RpcConnectioner, frameObj *DataFrame) {
		this.invokeAfterSendHandler(connObj, frameObj)
	})
	connObj.AddSendScheduleHandler("RpcServer.SendScheduleHandler", func(connObj RpcConnectioner) {
		this.invokeSendScheduleHandler(connObj)
	})
	connObj.AddBeforeHandleFrameHandler("RpcServer.BeforeHandleFrameHandler", func(connObj RpcConnectioner, frameObj *DataFrame) {
		this.invokeBeforeHandleFrameHandler(connObj, frameObj)
	})
	connObj.AddAfterInvokeHandler("RpcServer.AfterInvokeHandler", func(connObj RpcConnectioner, returnBytes []byte, err error) {
		this.invokeAfterInvokeHandler(connObj, returnBytes, err)
	})

	for _, item := range this.newConnectionHandlerData {
		item(connObj)
	}

	// 添加到连接集合中
	func() {
		this.connDataLockObj.Lock()
		defer this.connDataLockObj.Unlock()

		this.connData[connObj.ConnectionId()] = connObj
	}()
}

func (this *RpcServer) AddNewConnectionHandler(funcName string, funcObj func(connObj RpcConnectioner) error) (err error) {
	if _, exist := this.newConnectionHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.newConnectionHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcServer) onConnectionClose(connObj *RpcConnection4Server) {
	this.connDataLockObj.Lock()
	defer this.connDataLockObj.Unlock()

	delete(this.connData, connObj.ConnectionId())
}

func (this *RpcServer) Start(addr string, getConvertorFunc func() IByteConvertor) error {
	this.getConvertorFunc = getConvertorFunc
	defer log.Info("listen over Addr:%v", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("listen error Addr:%v error:%v", addr, err.Error())
		return err
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Error("Accept Error Addr:%v error:%v", listener.Addr(), err.Error())
			return err
		}

		rpcConnObj := NewRpcConnection4Server(con, this.ApiMgr, this.getConvertorFunc)
		this.invokeNewConnectionHandler(rpcConnObj)
	}
}

func (this *RpcServer) Start2(listener net.Listener, getConvertorFunc func() IByteConvertor) {
	this.getConvertorFunc = getConvertorFunc
	defer log.Info("listen over Addr:%v", listener.Addr())

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Error("Accept Error Addr:%v error:%v", listener.Addr(), err.Error())
			return
		}

		rpcConnObj := NewRpcConnection4Server(con, this.ApiMgr, this.getConvertorFunc)
		this.invokeNewConnectionHandler(rpcConnObj)
	}
}

func NewRpcServer() *RpcServer {
	result := &RpcServer{
		connData:                 make(map[int64]*RpcConnection4Server, 8),
		ApiMgr:                   newApiMgr(),
		RpcWatchBase:             newRpcWatchBase(),
		newConnectionHandlerData: make(map[string]func(connObj RpcConnectioner) error, 8),
	}

	return result
}
