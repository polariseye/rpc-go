package main

import (
	"fmt"
	"time"

	"github.com/polariseye/rpc-go"
	"github.com/polariseye/rpc-go/log"
)

var rpcObj = rpc.NewRpcClient(rpc.GetJsonConvertor)

func main() {
	// 注册客户端的服务
	rpcObj.RegisterService(new(Sample))

	rpcObj.AddAfterSendHandler("main", func(connObj rpc.RpcConnectioner, frameObj *rpc.DataFrame) {
		log.Debug("发送帧 Frame: RequestId:%d ResponseFrameId:%d ContentLength:%d TransformType:0X%x MethodName:%s",
			frameObj.RequestFrameId, frameObj.ResponseFrameId, frameObj.ContentLength, frameObj.TransformType(), frameObj.MethodName())
	})

	rpcObj.Start("127.0.0.1:50001", true)

	defer rpcObj.Close()

	callTst(rpcObj)
	time.Sleep(1000 * time.Second)
}

func callTst(rpcObj rpc.RpcConnectioner) {
	log.Debug("开始同步调用测试")
	defer log.Debug("完成同步调用测试")

	var err error
	var result = ""
	err = rpcObj.Call("global_Hello", []interface{}{"qqnihao"}, []interface{}{&result})
	if err != nil {
		fmt.Println("global_Hello 错误信息:", err.Error())
	} else {
		fmt.Println("global_Hello : 应答数据:", result)
	}

	err = rpcObj.Call("Sample_VoidTst", nil, nil)
	if err != nil {
		fmt.Println("Sample_VoidTst 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_VoidTst : 调用完成:")
	}

	err = rpcObj.Call("Sample_StringTst2", []interface{}{"qqnihao"}, []interface{}{&result})
	if err != nil {
		fmt.Println("Sample_StringTst2 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_StringTst2 : 应答数据:", result)
	}

	var result2 string
	err = rpcObj.Call("Sample_StringTst3", []interface{}{"qqnihao1", "qqnihao2"}, []interface{}{&result, &result2})
	if err != nil {
		fmt.Println("Sample_StringTst3 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_StringTst3 : 应答数据1:", result, " 应答数据2：", result2)
	}

	var manObj Man
	err = rpcObj.Call("Sample_StructTst1", []interface{}{Man{
		Name: "name1",
		Sex:  1,
	}}, []interface{}{&manObj})
	if err != nil {
		fmt.Println("Sample_StructTst1 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_StructTst1 : Name:", manObj.Name, " Sex:", manObj.Sex)
	}

	err = rpcObj.Call("Sample_CallClient", nil, nil)
	if err != nil {
		fmt.Println("Sample_CallClient 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_CallClient : 应答数据1:", result, " 应答数据2：", result2)
	}
}

func allTst(rpcObj *rpc.RpcConnection) {
	log.Debug("开始allTst测试")
	defer log.Debug("完成allTst测试")

	var err error
	var result = ""
	var waitChan <-chan error
	waitChan, err = rpcObj.CallAsync("global_Hello", []interface{}{"qqnihao"}, []interface{}{&result})
	if err != nil {
		fmt.Println("global_Hello CallAsync 错误信息:", err.Error())
	}
	err = <-waitChan
	if err != nil {
		fmt.Println("2 global_Hello CallAsync 错误信息:", err.Error())
	} else {
		fmt.Println("global_Hello CallAsync : 应答数据:", result)
	}

	err = rpcObj.CallAsyncWithNoResponse("global_Hello", []interface{}{"qqnihao"}, []interface{}{&result})
	if err != nil {
		fmt.Println("global_Hello CallAsyncWithNoResponse 错误信息:", err.Error())
	} else {
		fmt.Println("global_Hello CallAsyncWithNoResponse : 应答数据:", result)
	}

	err = rpcObj.CallTimeout("Sample_TimeoutTst", []interface{}{"qqnihao"}, []interface{}{&result}, 3*1000)
	if err != nil {
		fmt.Println("Sample_TimeoutTst CallTimeout 错误信息:", err.Error())
	} else {
		fmt.Println("Sample_TimeoutTst CallTimeout : 应答数据:", result)
	}
}
