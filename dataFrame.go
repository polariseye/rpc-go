package rpc

import "encoding/binary"

// 数据帧
type DataFrame struct {
	RequestObj      *RequestInfo
	Flag            byte
	RequestFrameId  uint32 //// 请求的帧Id
	ResponseFrameId uint32 //// 传输帧Id
	ContentLength   uint32 //// 内容长度
	MethodNameBytes []byte //// 方法名
	MethodNameLen   byte   ///// 方法名长度
	Data            []byte //// 内容具体数据
}

//// 传输类型 0:请求 1：应答
func (this *DataFrame) TransformType() byte {
	return this.Flag & 0x03
}

//// 是否是异常
func (this *DataFrame) IsError() bool {
	return this.Flag&0x04 == 0x04
}

func (this *DataFrame) SetError(errMsg string) {
	this.Data = []byte(errMsg)
	this.ContentLength = uint32(len(this.Data))
}

func (this *DataFrame) IsNeedResponse() bool {
	return this.Flag&0x08 == 0x08
}

func (this *DataFrame) SetIsNeedResponse(isNeedResponse bool) {
	if isNeedResponse {
		this.Flag = this.Flag | 0x08
	} else {
		this.Flag = this.Flag &^ 0x08 //// 把第四位置零
	}
}

func (this *DataFrame) SetData(data []byte) {
	this.MethodNameBytes = data[:this.MethodNameLen]
	this.Data = data[this.MethodNameLen:]
}

func (this *DataFrame) GetHeader(order binary.ByteOrder) []byte {
	header := make([]byte, HEADER_LENGTH)

	header[0] = HEADER
	header[1] = this.Flag
	header[HEADER_LENGTH-1] = TAIL
	order.PutUint32(header[2:], this.RequestFrameId)
	order.PutUint32(header[6:], this.ResponseFrameId)
	order.PutUint32(header[10:], this.ContentLength)
	header[15] = this.MethodNameLen

	return header
}

func (this *DataFrame) MethodName() string {
	return string(this.MethodNameBytes)
}

func convertHeader(header []byte, order binary.ByteOrder) *DataFrame {
	frameData := &DataFrame{}

	frameData.Flag = header[1]
	frameData.RequestFrameId = order.Uint32(header[2:4])
	frameData.ResponseFrameId = order.Uint32(header[6:9])

	frameData.MethodNameLen = header[15]
	frameData.ContentLength = order.Uint32(header[10:14])

	return frameData
}

func newRequestFrame(requestObj *RequestInfo, methodName string, data []byte, requestId uint32, isNeedResponse bool) *DataFrame {
	result := &DataFrame{
		RequestObj:      requestObj,
		Flag:            0,
		RequestFrameId:  requestId,
		ResponseFrameId: 0,
		MethodNameBytes: []byte(methodName),
		Data:            data,
	}

	result.MethodNameLen = byte(len(result.MethodNameBytes))
	result.ContentLength = uint32(len(data))
	result.SetIsNeedResponse(isNeedResponse)

	return result
}

func newResponseFrame(requestFrame *DataFrame, responseBytes []byte, requestFrameId uint32) *DataFrame {
	result := &DataFrame{
		Flag:            requestFrame.Flag,
		RequestFrameId:  requestFrameId,
		ResponseFrameId: requestFrame.RequestFrameId,
		ContentLength:   uint32(len(responseBytes)),
		MethodNameBytes: requestFrame.MethodNameBytes,
		MethodNameLen:   requestFrame.MethodNameLen,
		Data:            responseBytes,
	}

	return result
}
