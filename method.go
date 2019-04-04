package rpc

import (
	"reflect"

	"github.com/polariseye/rpc-go/log"
)

type MethodInfo struct {
	MethodName      string
	FuncObj         reflect.Value
	funcParamList   []reflect.Type
	returnValueList []reflect.Type
}

// 获取接口调用的参数
func (this *MethodInfo) GetInvokeParamList(connObj *RpcConnection, convertor IByteConvertor, data []byte) ([]reflect.Value, error) {
	var valList []reflect.Value
	if len(this.funcParamList) > 1 {
		var err error
		valList, err = convertor.UnMarhsalType(data, this.funcParamList[1:]...)
		if err != nil {
			log.Error("GetInvokeParamList error ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
			return nil, err
		}
	}

	// 组装请求数据
	var callValList = make([]reflect.Value, len(this.funcParamList))
	callValList[0] = reflect.ValueOf(connObj)
	for i := 0; i < len(valList); i++ {
		callValList[i+1] = valList[i]
	}

	return callValList, nil
}

// 接口调用
func (this *MethodInfo) Invoke(connObj *RpcConnection, paramList []reflect.Value) (responseValue []reflect.Value, err error) {
	defer func() {
		if tmpErr := recover(); tmpErr != nil {
			err = tmpErr.(error)
			log.Fatal("method call panic ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
		}
	}()

	responseValue = this.FuncObj.Call(paramList)
	//// 如果存在错误，则直接返回错误
	/*
		errVal := returnValList[len(returnValList)-1]
		if errVal.IsNil() == false {
			errObj := errVal.Interface().(error)
			return nil, errObj
		}
	*/

	// responseValue=responseValue[:len(returnValList)-1] //// 如果要支持error 需要把这儿打开
	return
}

// 获取应答字节数据
func (this *MethodInfo) GetResponseBytes(connObj *RpcConnection, responseList []reflect.Value, convertor IByteConvertor) (bytesResult []byte, err error) {
	bytesResult, err = convertor.MarshalType(this.returnValueList, responseList...)
	if err != nil {
		log.Error("GetResponseBytes error ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
		return
	}

	return
}

func newMethodInfo(methodName string, funcObj reflect.Value, paramList []reflect.Type, returnValList []reflect.Type) *MethodInfo {
	return &MethodInfo{
		MethodName:      methodName,
		FuncObj:         funcObj,
		funcParamList:   paramList,
		returnValueList: returnValList,
	}
}
