package main

import (
	"encoding/binary"
	"fmt"

	rpc "github.com/polariseye/rpc-go"
	"github.com/polariseye/rpc-go/protobufConvertor"
)

func main() {
	protobufConvertor.InitDefaultConvertor(binary.LittleEndian)
	serverObj := rpc.NewRpcServer(binary.LittleEndian, protobufConvertor.GetProtobufConvertor)

	serverObj.RegisterFunc("global", "hello", func(connObj *rpc.RpcConnection, personObj1 *Person, personObj2 *Person) (*Person, *Person) {
		personObj1.Id += 1
		personObj1.Name += "_____"
		personObj1.Phones = append(personObj1.Phones, "33")

		personObj2.Id += 1
		personObj2.Name += "_____"
		personObj2.Phones = append(personObj1.Phones, "33")

		return personObj1, personObj2
	})

	err := serverObj.Start("127.0.0.1:10001")
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
}
