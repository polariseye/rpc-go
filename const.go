package rpc

import (
	"errors"
	"reflect"
)

const (
	// 协议头字节
	HEADER = 0X09

	// 协议尾
	TAIL = 0x12

	// 协议头字节数
	HEADER_LENGTH = 16
)

// Flag信息
const (
	// 请求帧
	TransformType_Request = 0

	// 应答帧
	TransformType_Response = 1
)

var (
	RpcConnectionType = reflect.TypeOf((*RpcConnection)(nil))
	ErrorType         = reflect.TypeOf(error(nil))
)

var (
	TimeoutError = errors.New("timeout")
)

const (
	Yes = 1
	No  = 0
)
