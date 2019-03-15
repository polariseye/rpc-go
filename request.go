package rpc

import (
	"github.com/golang/protobuf/proto"
)

type RequestLink struct {
}

type RequestNode struct {
	PreNode  *RequestNode
	NextNode *RequestNode
	Disposed bool
	Data     *RequestInfo
}

type RequestInfo struct {
	IsResponsed bool
	RequestId   uint32 //// 请求Id
	DownChan    chan error

	ReturnObj   proto.Message
	ErrObj      error
	ReturnBytes []byte
}

/*
暂存的帧链表
实例池
*/
