package rpc

import (
	"sync"
	"sync/atomic"
	"time"
)

type RequestInfo struct {
	IsResponsed int32
	RequestId   uint32 //// 请求Id
	DownChan    chan error

	ReturnObj   []interface{}
	ErrObj      error
	ReturnBytes []byte

	// 过期时间点
	ExpireTime int64
}

func (this *RequestInfo) Return(returnObj []interface{}, returnBytes []byte, err error) bool {
	if atomic.CompareAndSwapInt32(&this.IsResponsed, No, Yes) == false {
		return false
	}

	this.ReturnObj = returnObj
	this.ReturnBytes = returnBytes
	this.ErrObj = err
	this.DownChan <- err

	return true
}

func (this *RequestInfo) ReturnError(err error) bool {
	if atomic.CompareAndSwapInt32(&this.IsResponsed, No, Yes) == false {
		return false
	}

	this.ErrObj = err
	this.DownChan <- err

	return true
}

type FrameContainer struct {
	data         map[uint32]*RequestInfo
	preCheckTime int64

	lockObj sync.RWMutex
}

// 获取并删除项
func (this *FrameContainer) GetRequestInfo(frameId uint32) (result *RequestInfo, exist bool) {
	this.lockObj.RLock()
	defer this.lockObj.RUnlock()

	result, exist = this.data[frameId]

	return result, exist
}

func (this *FrameContainer) AddRequest(requestObj *RequestInfo) {
	this.lockObj.Lock()
	defer this.lockObj.Unlock()

	this.data[requestObj.RequestId] = requestObj
}

func (this *FrameContainer) RemoveRequestObj(requestId uint32) {
	this.lockObj.Lock()
	defer this.lockObj.Unlock()

	delete(this.data, requestId)
}

func (this *FrameContainer) ClearExpireNode() {
	now := time.Now().Unix()

	// 每秒检查一次
	if (now - this.preCheckTime) < 1 {
		return
	}
	this.preCheckTime = now

	// 查找过期节点
	var expireNode []*RequestInfo = nil
	func() {
		this.lockObj.RLock()
		defer this.lockObj.RUnlock()

		for _, item := range this.data {
			if item.ExpireTime < now {
				if expireNode == nil {
					expireNode = make([]*RequestInfo, 0, 8)
				}
				if item.ReturnError(TimeoutError) {
					expireNode = append(expireNode, item)
				}
			}
		}
	}()

	if expireNode == nil {
		return
	}

	// 删除过期节点
	this.lockObj.Lock()
	defer this.lockObj.Unlock()

	for _, item := range expireNode {
		delete(this.data, item.RequestId)
	}

	return
}

func (this *FrameContainer) ReturnAllRequest(err error) {
	this.lockObj.Lock()
	defer this.lockObj.Unlock()

	for _, item := range this.data {
		item.ReturnError(err)
	}

	// 清空所有
	this.data = make(map[uint32]*RequestInfo, 16)
}

func newFrameContainer() *FrameContainer {
	result := &FrameContainer{
		data:    make(map[uint32]*RequestInfo, 16),
		lockObj: sync.RWMutex{},
	}

	return result
}
