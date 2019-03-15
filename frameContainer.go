package rpc

type FrameContainer struct {
}

// 获取并删除项
func (this *FrameContainer) GetRequestInfo(frameId uint32) (result *RequestInfo, exist bool) {
	return nil, false
}

func (this *FrameContainer) AddRequest(requestObj *RequestInfo) {

}

func (this *FrameContainer) RemoveRequestObj(requestId uint32) {

}
