package protoConvertor

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
		if item == nil {
			byteData = append(byteData, 0x00, 0x00, 0x00, 0x00)
			continue
		}

		pbItem := item.(Protobuffer)
		if pbItem == nil {
			byteData = append(byteData, 0x00, 0x00, 0x00, 0x00)
			continue
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
		data = append(data, item.Interface())
	}

	return this.MarshalValue(data...)
}

func (this *ProtobufConvertor) UnMarhsalType(bytesData []byte, typeList ...reflect.Type) ([]reflect.Value, error) {
	data := make([]interface{}, 0, len(typeList))
	result := make([]reflect.Value, 0, len(typeList))
	for _, item := range typeList {
		valItem := reflect.New(item)
		data = append(data, valItem.Interface())
		valItem = reflect.Indirect(valItem)
		result = append(result, valItem)
	}

	return nil, this.UnMarhsalValue(bytesData, data...)
}

func (this *ProtobufConvertor) UnMarhsalValue(bytesData []byte, valList ...interface{}) error {
	var handledLen uint32
	var dataLen uint32
	for _, valItem := range valList {
		dataLen = this.byteOrder.Uint32(bytesData[handledLen : handledLen+4])
		pbItem := valItem.(Protobuffer)

		handledLen += 4

		err := pbItem.Unmarshal(bytesData[handledLen : handledLen+dataLen])
		if err != nil {
			return err
		}

		handledLen += dataLen
	}

	return nil
}

var protoConvertorObj = new(ProtobufConvertor)

func GetProtobufConvertor() rpc.IByteConvertor {
	return protoConvertorObj
}
