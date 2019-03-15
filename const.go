package rpc

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
