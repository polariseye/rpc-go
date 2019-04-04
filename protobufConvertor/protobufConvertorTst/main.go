package main

import (
	"encoding/binary"
	"fmt"

	"github.com/polariseye/rpc-go/protobufConvertor"
)

func main() {
	ValueTst()
}

func ValueTst() {
	var pList = make([]interface{}, 0)
	pList = append(pList, Person{
		Id:     1,
		Name:   "今天天气好坏啊",
		Phones: []string{"1", "2"},
	}, Person{
		Id:     2,
		Name:   "今天天气好坏啊2",
		Phones: []string{"12", "22"},
	})

	protobufConvertor.InitDefaultConvertor(binary.LittleEndian)
	convertorObj := protobufConvertor.GetProtobufConvertor()
	data, err := convertorObj.MarshalValue(pList...)
	if err != nil {
		fmt.Println("MarshalValue error:%v", err.Error())
		return
	}

	var resultp1 Person
	var resultp2 Person
	err = convertorObj.UnMarhsalValue(data, &resultp1, &resultp2)
	if err != nil {
		fmt.Println("UnMarhsalValue error:%v", err.Error())
		return
	}

	fmt.Println("result:", resultp1.String(), "result2", resultp2.String())
}
