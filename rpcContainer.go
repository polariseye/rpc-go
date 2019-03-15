package rpc

import "net"

type RpcContainer struct {
	funcData map[string]interface{}
}

// 注册一个RPC服务端
func (this *RpcContainer) RegisterService(obj interface{}) {

}

func (this *RpcContainer) RegisterFunc() {

}

func (this *RpcContainer) GetRpcConnection(con net.Conn) {

}

func (this *RpcContainer) getMethod(methodName string) (*MethodInfo, bool) {

}

func NewRpcContainer() *RpcContainer {
	return &RpcContainer{
		funcData: make(map[string]interface{}),
	}
}
