package rpc

import (
	"encoding/binary"
	"reflect"

	"github.com/polariseye/rpc-go/log"
)

type MethodInfo struct {
	MethodName      string
	FuncObj         reflect.Value
	funcParamList   []reflect.Type
	returnValueList []reflect.Type
}

func (this *MethodInfo) Invoke(connObj *RpcConnection, convertor IByteConvertor, data []byte, order binary.ByteOrder, ifNeedResponse bool) (bytesResult []byte, err error) {
	defer func() {
		if tmpErr := recover(); tmpErr != nil {
			err = tmpErr.(error)
			log.Fatal("method call panic ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
		}
	}()

	var valList []reflect.Value
	if len(this.funcParamList) > 1 {
		valList, err = convertor.UnMarhsalType(data, this.funcParamList[1:]...)
		if err != nil {
			log.Error("UnMarhsalType error ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
			return nil, err
		}
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

	if ifNeedResponse {
		if len(this.returnValueList) > 0 {
			// bytesResult, err = convertor.MarshalType(this.returnValueList, returnValList[:len(returnValList)]...) //// 如果要支持error 需要把这儿打开
			bytesResult, err = convertor.MarshalType(this.returnValueList, returnValList...)
			if err != nil {
				log.Error("MarshalType error ip:%v MethodName:%v error:%v", connObj.Addr(), this.MethodName, err.Error())
				return
			}
		}
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
