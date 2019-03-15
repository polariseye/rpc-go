package rpc

import (
	"encoding/binary"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type MethodInfo struct {
	MethodName      string
	FuncObj         *reflect.Value
	funcParamList   []reflect.Type
	returnValueList []reflect.Type
}

type sizer interface {
	XXX_Size() int
}

func (this *MethodInfo) Invoke(data []byte, order binary.ByteOrder) ([]byte, error) {
	valList := make([]reflect.Value, len(this.funcParamList))

	var startByteIndex int = 0
	var valItem reflect.Value
	var err error
	for i := 0; i < len(this.funcParamList); i++ {
		valItem, startByteIndex, err = this.adjustParam(this.funcParamList[i], data[startByteIndex:])
		if err != nil {
			return nil, err
		}

		valList[i] = valItem
	}

	returnValList := this.FuncObj.Call(valList)

	return this.convertToBytes(returnValList), nil
}

func (this *MethodInfo) adjustParam(tp reflect.Type, data []byte) (reflect.Value, int, error) {
	if tp.Kind() == reflect.Array {
		return reflect.ValueOf(data), len(data), nil
	} else {
		val := reflect.New(tp)
		obj := val.Interface().(proto.Message)
		err := proto.Unmarshal(data, obj)

		return val, obj.(sizer).XXX_Size(), err
	}
}

func (this *MethodInfo) convertToBytes(valList []reflect.Value) []byte {
	return nil
}
