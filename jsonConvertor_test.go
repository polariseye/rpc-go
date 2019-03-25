package rpc

import (
	"fmt"
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	convertor := JsonConvertor{}

	bytesData, err := convertor.MarshalType(nil, reflect.ValueOf(1))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("json数据:", string(bytesData))

	val, err := convertor.UnMarhsalType(bytesData, reflect.TypeOf(1))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("\r\n:%v", val)
}

func TestValue(t *testing.T) {
	convertor := JsonConvertor{}

	bytesData, err := convertor.MarshalValue(1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("json数据:", string(bytesData))

	val := make([]interface{}, 1)
	tmpVal := 0
	val[0] = &tmpVal
	err = convertor.UnMarhsalValue(bytesData, val...)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("\r\n:%v", val)
}
