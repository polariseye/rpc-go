package rpc

import (
	"encoding/json"
	"reflect"
)

type JsonConvertor struct {
}

func (this *JsonConvertor) MarshalValue(valList ...interface{}) ([]byte, error) {
	bytesData, err := json.Marshal(valList)

	return bytesData, err
}

func (this *JsonConvertor) MarshalType(typeList []reflect.Type, valList ...reflect.Value) ([]byte, error) {
	data := make([]interface{}, 0, len(valList))
	for _, item := range valList {
		data = append(data, item.Interface())
	}

	bytesData, err := json.Marshal(data)
	return bytesData, err
}

func (this *JsonConvertor) UnMarhsalType(bytesData []byte, typeList ...reflect.Type) ([]reflect.Value, error) {
	data := make([]interface{}, 0, len(typeList))
	result := make([]reflect.Value, 0, len(typeList))
	for _, item := range typeList {
		valItem := reflect.New(item)
		data = append(data, valItem.Interface())
		valItem = reflect.Indirect(valItem)
		result = append(result, valItem)
	}

	err := json.Unmarshal(bytesData, &data)

	return result, err
}

func (this *JsonConvertor) UnMarhsalValue(bytesData []byte, valList ...interface{}) error {
	err := json.Unmarshal(bytesData, &valList)
	return err
}

var jsonConvertorObj = new(JsonConvertor)

func GetJsonConvertor() IByteConvertor {
	return jsonConvertorObj
}
