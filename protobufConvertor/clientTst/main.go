package main

import (
	"encoding/binary"
	"fmt"

	rpc "github.com/polariseye/rpc-go"
	"github.com/polariseye/rpc-go/protobufConvertor"
)

func main() {
	protobufConvertor.InitDefaultConvertor(binary.BigEndian)
	rpcClientObj := rpc.NewRpcClient(protobufConvertor.GetProtobufConvertor)

	err := rpcClientObj.Start("127.0.0.1:10001", true)
	if err != nil {
		fmt.Println("connect error:", err.Error())
		return
	}

	var pList = make([]interface{}, 0)
	pList = append(pList, &Person{
		Id:     1,
		Name:   "今天天气好坏啊",
		Phones: []string{"1", "2"},
	}, &Person{
		Id:     2,
		Name:   "今天天气好坏啊2",
		Phones: []string{"12", "22"},
	})

	var resultList = make([]interface{}, 0)
	resultList = append(resultList, &Person{}, &Person{})

	err = rpcClientObj.Call("global_hello", pList, resultList)
	if err != nil {
		fmt.Println("call error:", err.Error())
		return
	}

	fmt.Println(resultList)
}
