package rpc

import (
	"net"
	"sync"
	"time"

	"github.com/polariseye/rpc-go/log"
)

// 客户端连接对象
type RpcClient struct {
	*RpcConnection4Client
	*ApiMgr

	isAutoReconnect  bool
	addr             string
	getConvertorFunc func() IByteConvertor

	isStopped            *bool //// 用指针是为了避免在调用Start时，正在进行重连
	autoReconnectLockObj sync.Mutex
}

// 关闭连接
// 如果重复调用，将会只生效一次
func (this *RpcClient) Close() {
	*this.isStopped = true
	this.isStopped = new(bool)

	this.RpcConnection4Client.Close()
}

// Start 连接到指定地址
// addr: 服务端地址
// isAutoReconnect: 是否自动重连到服务端
// getConvertorFunc: 转换对象获取函数（协议处理用）
// 返回值:
// error:错误信息
func (this *RpcClient) Start(addr string, isAutoReconnect bool, getConvertorFunc func() IByteConvertor) error {
	if *this.isStopped == false {
		return HaveConnectedError
	}
	if this.IsClosed() == false {
		return HaveConnectedError
	}

	// 因为可能重连时会用到isStopped,确保重连协程能够正常退出，所以new一个新对象
	*this.isStopped = true
	this.isStopped = new(bool)

	this.autoReconnectLockObj.Lock()
	defer this.autoReconnectLockObj.Unlock()

	this.isAutoReconnect = isAutoReconnect
	this.getConvertorFunc = getConvertorFunc
	this.addr = addr

	if isAutoReconnect {
		// 开启重连
		go this.reconnect(this.isStopped)
	} else {
		// 只连接一次
		if this.connect(this.isStopped, addr) {
			return nil
		}

		return ConnectionTimeOut
	}

	return nil
}

// Start2 使用指定连接进行协议处理
// 使用此函数开启处理进，将不会进行断线重连。连接完全由外部处理
// con:连接对象
// getConvertorFunc:转换对象获取函数（协议处理用）
// 返回值:
// error:错误信息
func (this *RpcClient) Start2(con net.Conn, getConvertorFunc func() IByteConvertor) error {
	if *this.isStopped == false {
		return HaveConnectedError
	}
	if this.IsClosed() == false {
		return HaveConnectedError
	}

	// 因为可能重连时会用到isStopped,确保重连协程能够正常退出，所以new一个新对象
	*this.isStopped = true
	this.isStopped = new(bool)

	this.autoReconnectLockObj.Lock()
	defer this.autoReconnectLockObj.Unlock()
	this.getConvertorFunc = getConvertorFunc
	conObj := newRpcConnection(this.ApiMgr, con, this, getConvertorFunc)
	this.RpcConnection4Client.setConnection(conObj)

	return nil
}

// Addr 获取服务端地址
// 如果没有连接信息，则会返回空字符串
func (this *RpcClient) Addr() string {
	if this.addr != "" {
		return this.addr
	}

	if this.RpcConnection.IsClosed() == false {
		return this.RpcConnection4Client.Addr()
	}

	return ""
}

// reconnect 重连
// isStopped:用于判断是否已经停止重连了，使用指针是为了避免调用Start导致多个重连协程的问题
func (this *RpcClient) reconnect(isStopped *bool) {
	if this.isAutoReconnect == false {
		log.Debug("no need auto reconnect")
		return
	}

	addr := this.addr
	for *isStopped == false && this.isAutoReconnect { //// 地址有变更，则立即停止重连
		if this.connect(isStopped, addr) {
			break
		}

		time.Sleep(5 * time.Second)
	}

}

// connect
func (this *RpcClient) connect(isStopped *bool, addr string) bool {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		log.Debug("fail to connect to server addr:%v error:%v", addr, err.Error())
		return false

	}

	this.autoReconnectLockObj.Lock()
	defer this.autoReconnectLockObj.Unlock()
	if *isStopped {
		con.Close()
	}

	conObj := newRpcConnection(this.ApiMgr, con, this, this.getConvertorFunc)
	this.RpcConnection4Client.setConnection(conObj)

	return true
}

// NewRpcClient 新建Rpc连接客户端对象
func NewRpcClient() *RpcClient {
	result := &RpcClient{
		ApiMgr:          newApiMgr(),
		isAutoReconnect: false,
		isStopped:       new(bool),
	}

	// 添加对自动重连的支持
	result.AddCloseHandler("RpcClient.reconnect", func(conObj RpcConnectioner) {
		if result.isAutoReconnect {
			go result.reconnect(result.isStopped)
		}

		return
	})

	return result
}
