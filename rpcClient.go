package rpc

import (
	"encoding/binary"
	"net"
	"sync"

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
	byteOrder            binary.ByteOrder
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
// 返回值:
// error:错误信息
func (this *RpcClient) Start(addr string, isAutoReconnect bool) error {
	var result error
	func() {
		this.autoReconnectLockObj.Lock()
		defer this.autoReconnectLockObj.Unlock()

		if *this.isStopped == false {
			result = HaveConnectedError
			return
		}
		if this.IsClosed() == false {
			result = HaveConnectedError
			return
		}

		// 因为可能重连时会用到isStopped,确保重连协程能够正常退出，所以new一个新对象
		*this.isStopped = true
		this.isStopped = new(bool)

		this.isAutoReconnect = isAutoReconnect
		this.addr = addr
	}()

	if result != nil {
		return result
	}

	// 先尝试连接一次
	log.Info("start reconnect to %v", addr)
	if this.connect(this.isStopped, addr) {
		return nil
	}

	if isAutoReconnect {
		// 开启重连
		go this.reconnect(this.isStopped)
	} else {
		return ConnectionTimeOut
	}

	return nil
}

// Start2 使用指定连接进行协议处理
// 使用此函数开启处理进，将不会进行断线重连。连接完全由外部处理
// con:连接对象
// 返回值:
// error:错误信息
func (this *RpcClient) Start2(con net.Conn) error {
	this.autoReconnectLockObj.Lock()
	defer this.autoReconnectLockObj.Unlock()

	if *this.isStopped == false {
		return HaveConnectedError
	}
	if this.IsClosed() == false {
		return HaveConnectedError
	}

	// 因为可能重连时会用到isStopped,确保重连协程能够正常退出，所以new一个新对象
	*this.isStopped = true
	this.isStopped = new(bool)

	conObj := newRpcConnection(this.ApiMgr, con, this, this.byteOrder, this.getConvertorFunc)
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
		log.Info("start reconnect to %v", addr)
		if this.connect(isStopped, addr) {
			break
		}
	}
}

// connect
func (this *RpcClient) connect(isStopped *bool, addr string) bool {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		log.Info("fail to connect to server addr:%v error:%v", addr, err.Error())
		return false
	}

	this.autoReconnectLockObj.Lock()
	defer this.autoReconnectLockObj.Unlock()
	if *isStopped {
		log.Info("change server old server:%v", addr)
		con.Close()
	}

	conObj := newRpcConnection(this.ApiMgr, con, this, this.byteOrder, this.getConvertorFunc)
	this.RpcConnection4Client.setConnection(conObj)
	log.Info("connected to server:%v", addr)

	return true
}

// NewRpcClient 新建Rpc连接客户端对象
// getConvertorFunc:转换对象获取函数（协议处理用）
func NewRpcClient(byteOrder binary.ByteOrder, getConvertorFunc func() IByteConvertor) *RpcClient {
	result := &RpcClient{
		ApiMgr:               newApiMgr(),
		isAutoReconnect:      false,
		isStopped:            new(bool),
		getConvertorFunc:     getConvertorFunc,
		RpcConnection4Client: NewRpcConnection4Client(),
		byteOrder:            byteOrder,
	}

	*result.isStopped = true

	// 添加对自动重连的支持
	result.AddCloseHandler("RpcClient.reconnect", func(conObj RpcConnectioner) {
		if result.isAutoReconnect {
			go result.reconnect(result.isStopped)
		}

		return
	})

	return result
}
