package main

import (
	"fmt"
	"net"

	"github.com/polariseye/rpc-go"
)

var containerObj = rpc.NewRpcContainer(rpc.GetJsonConvertor)

func main() {
	con, err := net.Dial("tcp", "127.0.0.1:50001")
	if err != nil {
		fmt.Println("出错：", err.Error())
		return
	}

	rpcObj := containerObj.GetRpcConnection(con)
	defer rpcObj.Close(nil)

	var result = ""
	err = rpcObj.Call("global_Hello", &result, "qqnihao")
	if err != nil {
		fmt.Println("错误信息:", err.Error())
		return
	}

	fmt.Println("应答数据:", result)
}
