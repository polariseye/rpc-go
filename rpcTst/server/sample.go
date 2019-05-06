package main

import (
	"fmt"
	"time"

	"github.com/polariseye/rpc-go"
)

type Sample struct {
}

func (this *Sample) VoidTst(connObj rpc.RpcConnectioner) {
	fmt.Println("调用成功!!!!!!!!!!!!!!!!!!!")
	return
}

func (this *Sample) StringTst1(connObj rpc.RpcConnectioner) string {
	return "你好"
}

func (this *Sample) StringTst2(connObj rpc.RpcConnectioner, name string) string {
	return "你好:" + name
}

// 多参数多返回值
func (this *Sample) StringTst3(connObj rpc.RpcConnectioner, name string, name2 string) (string, string) {
	return "你好1：" + name, "你好2:" + name2
}

type Man struct {
	Name string
	Sex  int
}

func (this *Sample) StructTst1(connObj rpc.RpcConnectioner, man Man) (result Man) {
	man.Name = "Server" + man.Name
	man.Sex = 10 + man.Sex

	return man
}

func (this *Sample) CallClient(connObj rpc.RpcConnectioner) {
	responseString := ""
	err := connObj.Call("Sample_StringTst1", nil, []interface{}{&responseString})
	if err != nil {
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println("result:", responseString)
	}
}

func (this *Sample) TimeoutTst(connObj rpc.RpcConnectioner) {
	time.Sleep(5 * time.Second)
}
