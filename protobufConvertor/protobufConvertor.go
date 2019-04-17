package protobufConvertor

import (
	"encoding/binary"
	"reflect"

	"github.com/polariseye/rpc-go"
)

type ProtobufConvertor struct {
	byteOrder binary.ByteOrder
}

type Protobuffer interface {
	Marshal() (dAtA []byte, err error)
	Unmarshal(dAtA []byte) error
	Size() (n int)
}

func (this *ProtobufConvertor) MarshalValue(valList ...interface{}) ([]byte, error) {
	var lensBytes = make([]byte, 4)
	byteData := make([]byte, 0, 2040)
	for _, item := range valList {
		if rpc.IsNil(item) {
			len0 := uint32(0)
			this.byteOrder.PutUint32(lensBytes, len0)
			byteData = append(byteData, lensBytes...)
			continue
		}

		pbItem, ok := item.(Protobuffer)
		if ok == false {
			return nil, rpc.NotSupportedTypeError
		}

		var dataLen uint32 = 0
		tmpBytes, err := pbItem.Marshal()
		if err != nil {
			return nil, err
		}

		dataLen = uint32(len(tmpBytes))
		this.byteOrder.PutUint32(lensBytes, dataLen)

		byteData = append(byteData, lensBytes...)
		byteData = append(byteData, tmpBytes...)
	}

	return byteData, nil
}

func (this *ProtobufConvertor) MarshalType(typeList []reflect.Type, valList ...reflect.Value) ([]byte, error) {
	data := make([]interface{}, 0, len(valList))
	for _, item := range valList {
		if item.Kind() == reflect.Struct {
			val := item.Interface()
			data = append(data, val)
		} else {
			data = append(data, item.Interface())
		}
	}

	return this.MarshalValue(data...)
}

func (this *ProtobufConvertor) UnMarhsalType(bytesData []byte, typeList ...reflect.Type) ([]reflect.Value, error) {
	data := make([]interface{}, 0, len(typeList))
	result := make([]reflect.Value, 0, len(typeList))
	for _, item := range typeList {
		valItem := reflect.New(item)

		// 实例化指针指向的值
		tmpItem := item
		tmpVal := valItem
		for tmpItem.Kind() == reflect.Ptr {
			tmpItem = tmpItem.Elem()

			childVal := reflect.New(tmpItem)
			tmpVal.Elem().Set(childVal)

			tmpVal = childVal
		}

		data = append(data, valItem.Elem().Interface())
		result = append(result, valItem.Elem())
	}

	return result, this.UnMarhsalValue(bytesData, data...)
}

func (this *ProtobufConvertor) UnMarhsalValue(bytesData []byte, valList ...interface{}) error {
	var handledLen uint32
	var dataLen uint32
	for _, valItem := range valList {
		dataLen = this.byteOrder.Uint32(bytesData[handledLen : handledLen+4])
		handledLen += 4
		if dataLen == 0 {
			// 如果没有数据，则跳过这项
			continue
		}

		pbItem, ok := valItem.(Protobuffer)
		if ok == false {
			return rpc.NotSupportedTypeError
		}

		err := pbItem.Unmarshal(bytesData[handledLen : handledLen+dataLen])
		if err != nil {
			return err
		}

		handledLen += dataLen
	}

	return nil
}

func newProtobufConvertor(byteOrder binary.ByteOrder) *ProtobufConvertor {
	return &ProtobufConvertor{
		byteOrder: byteOrder,
	}
}

var (
	protoConvertorObj *ProtobufConvertor
)

func InitDefaultConvertor(byteOrder binary.ByteOrder) {
	protoConvertorObj = newProtobufConvertor(byteOrder)
}

func GetProtobufConvertor() rpc.IByteConvertor {
	return protoConvertorObj
}
