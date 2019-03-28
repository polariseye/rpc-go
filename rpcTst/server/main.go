package main

import (
	"fmt"
	"net"

	"github.com/polariseye/rpc-go"
)

var containerObj = rpc.NewRpcContainer(rpc.GetJsonConvertor)

func Hello(connObj *rpc.RpcConnection, name string) (say string, err error) {
	return fmt.Sprintf("你好哈:%v", name), nil
}

func main() {
	containerObj.RegisterFunc("global", "Hello", Hello)
	containerObj.RegisterService(new(Sample))

	containerObj.RecordAllMethod()

	listenObj, err := net.Listen("tcp", "127.0.0.1:50001")
	if err != nil {
		fmt.Println("出错：", err.Error())
		return
	}

	for {
		con, err := listenObj.Accept()
		if err != nil {
			fmt.Println("出错：", err.Error())
		}

		containerObj.GetRpcConnection(con)
	}
}
