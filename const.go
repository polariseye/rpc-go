package rpc

import (
	"errors"
	"reflect"
)

const (
	// 协议头字节
	HEADER byte = 0X09

	// 协议尾
	TAIL byte = 0x12

	// 协议头字节数
	HEADER_LENGTH = 16
)

// Flag信息
const (
	// 正常的数据帧
	TransformType_Nomal byte = 0x00

	// 心跳
	TransformType_KeepAlive byte = 0x01
)

var (
	RpcConnectionType = reflect.TypeOf((*RpcConnection)(nil))
	ErrorType         = reflect.TypeOf((*error)(nil)).Elem() //// 这里必须用指针，否则提示为Nil
)

var (
	CustCloseConnectionError = errors.New("CustCloseConnectionError")
	CallTimeoutError         = errors.New("CallTimeoutError")
	ConnectionTimeOut        = errors.New("ConnectionTimeOut")
	MethodNotFoundError      = errors.New("MethodNotFound")
	HandlerExistedError      = errors.New("HandlerExisted")
	HaveConnectedError       = errors.New("HaveConnectedError")
	NilError                 = errors.New("NilError")
	NotSupportedTypeError    = errors.New("NotSupportedTypeError")
	InnerDataError           = errors.New("InnerDataError")
	ConnectionClosedError    = errors.New("ConnectionClosedError")
)

const (
	Yes = 1
	No  = 0
)
