package rpc

import (
	"encoding/binary"
	"reflect"
)

type MethodInfo struct {
	MethodName      string
	FuncObj         reflect.Value
	funcParamList   []reflect.Type
	returnValueList []reflect.Type
}

func (this *MethodInfo) Invoke(connObj *RpcConnection, convertor IByteConvertor, data []byte, order binary.ByteOrder) ([]byte, error) {
	valList, err := convertor.UnMarhsalType(data, this.funcParamList[1:]...)
	if err != nil {
		return nil, err
	}

	// 组装请求数据
	var callValList = make([]reflect.Value, len(this.funcParamList))
	callValList[0] = reflect.ValueOf(connObj)
	for i := 0; i < len(valList); i++ {
		callValList[i+1] = valList[i]
	}

	returnValList := this.FuncObj.Call(callValList)
	//// 如果存在错误，则直接返回错误
	/*
		errVal := returnValList[len(returnValList)-1]
		if errVal.IsNil() == false {
			errObj := errVal.Interface().(error)
			return nil, errObj
		}
	*/

	return convertor.MarshalType(this.returnValueList, returnValList[:len(returnValList)-1]...)
}

func newMethodInfo(methodName string, funcObj reflect.Value, paramList []reflect.Type, returnValList []reflect.Type) *MethodInfo {
	return &MethodInfo{
		MethodName:      methodName,
		FuncObj:         funcObj,
		funcParamList:   paramList,
		returnValueList: returnValList,
	}
}
