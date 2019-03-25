package rpc

import "reflect"

type IByteConvertor interface {
	MarshalValue(valList ...interface{}) ([]byte, error)
	MarshalType(typeList []reflect.Type, valList ...reflect.Value) ([]byte, error)
	UnMarhsalType(bytesData []byte, typeList ...reflect.Type) ([]reflect.Value, error)
	UnMarhsalValue(bytesData []byte, valList ...interface{}) error
}
