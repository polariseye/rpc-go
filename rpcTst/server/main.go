package main

import (
	"fmt"
	"time"

	"github.com/polariseye/rpc-go"
	"github.com/polariseye/rpc-go/log"
)

var rpcServerObj = rpc.NewRpcServer(rpc.GetJsonConvertor)

func Hello(connObj *rpc.RpcConnection, name string) (say string, err error) {
	return fmt.Sprintf("你好哈:%v", name), nil
}

func main() {
	rpcServerObj.RegisterFunc("global", "Hello", Hello)
	rpcServerObj.RegisterService(new(Sample))

	rpcServerObj.RecordAllMethod()

	rpcServerObj.AddBeforeHandleFrameHandler("main", func(connObj rpc.RpcConnectioner, frameObj *rpc.DataFrame) {
		log.Debug("Frame: RequestId:%d ResponseFrameId:%d ContentLength:%d TransformType:0X%x MethodName:%s",
			frameObj.RequestFrameId, frameObj.ResponseFrameId, frameObj.ContentLength, frameObj.TransformType(), frameObj.MethodName())
	})

	go func() {
		for {
			fmt.Println("连接数量:", rpcServerObj.GetConnectionCount())
			time.Sleep(time.Millisecond * 200)
		}
	}()

	rpcServerObj.Start("127.0.0.1:50001")
}
