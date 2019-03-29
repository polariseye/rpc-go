package rpc

type RpcWatcher interface {
	afterSend(frameObj *DataFrame) (err error)
	sendSchedule() (err error)
	beforeHandleFrame(frameObj *DataFrame) (isHandled bool, err error)
	afterInvoke(frameObj *DataFrame, returnBytes []byte, err error)
	afterClose()
}

type RpcWatchBase struct {
	afterSendHandlerData         map[string]func(connObj RpcConnectioner, frameObj *DataFrame)
	closeHandlerData             map[string]func(connObj RpcConnectioner)
	sendScheduleHandlerData      map[string]func(connObj RpcConnectioner)
	beforeHandleFrameHandlerData map[string]func(connObj RpcConnectioner, frameObj *DataFrame)
	afterInvokeHandlerData       map[string]func(connObj RpcConnectioner, returnBytes []byte, err error)
}

func (this *RpcWatchBase) AddCloseHandler(funcName string, funcObj func(connObj RpcConnectioner)) (err error) {
	if _, exist := this.closeHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.closeHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcWatchBase) invokeCloseHandler(connObj RpcConnectioner) {
	for _, item := range this.closeHandlerData {
		item(connObj)
	}
}

func (this *RpcWatchBase) AddAfterSendHandler(funcName string, funcObj func(connObj RpcConnectioner, frameObj *DataFrame)) (err error) {
	if _, exist := this.afterSendHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.afterSendHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcWatchBase) invokeAfterSendHandler(connObj RpcConnectioner, frameObj *DataFrame) {
	for _, item := range this.afterSendHandlerData {
		item(connObj, frameObj)
	}
}

func (this *RpcWatchBase) AddSendScheduleHandler(funcName string, funcObj func(connObj RpcConnectioner)) (err error) {
	if _, exist := this.sendScheduleHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.sendScheduleHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcWatchBase) invokeSendScheduleHandler(connObj RpcConnectioner) {
	for _, item := range this.sendScheduleHandlerData {
		item(connObj)
	}
}

func (this *RpcWatchBase)AddBeforeHandleFrameHandler(funcName string, funcObj func(connObj RpcConnectioner, frameObj *DataFrame)) (err error) {
	if _, exist := this.beforeHandleFrameHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.beforeHandleFrameHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcWatchBase) invokeBeforeHandleFrameHandler(connObj RpcConnectioner, frameObj *DataFrame) {
	for _, item := range this.beforeHandleFrameHandlerData {
		item(connObj, frameObj)
	}
}

func (this *RpcWatchBase) AddAfterInvokeHandler(funcName string, funcObj func(connObj RpcConnectioner, returnBytes []byte, err error)) (err error) {
	if _, exist := this.afterInvokeHandlerData[funcName]; exist {
		return HandlerExistedError
	}

	this.afterInvokeHandlerData[funcName] = funcObj
	return nil
}

func (this *RpcWatchBase) invokeAfterInvokeHandler(connObj RpcConnectioner, returnBytes []byte, err error) {
	for _, item := range this.afterInvokeHandlerData {
		item(connObj, returnBytes, err)
	}
}

func newRpcWatchBase() *RpcWatchBase {
	return &RpcWatchBase{
		afterSendHandlerData:         make(map[string]func(connObj RpcConnectioner, frameObj *DataFrame), 4),
		closeHandlerData:             make(map[string]func(connObj RpcConnectioner), 4),
		sendScheduleHandlerData:      make(map[string]func(connObj RpcConnectioner), 4),
		beforeHandleFrameHandlerData: make(map[string]func(connObj RpcConnectioner, frameObj *DataFrame), 4),
		afterInvokeHandlerData:       make(map[string]func(connObj RpcConnectioner, returnBytes []byte, err error), 4),
	}
}
