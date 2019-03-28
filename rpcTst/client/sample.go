package main

import (
	"fmt"

	"github.com/polariseye/rpc-go"
)

type Sample struct {
}

func (this *Sample) VoidTst(connObj *rpc.RpcConnection) error {
	fmt.Println("调用成功!!!!!!!!!!!!!!!!!!!")
	return nil
}

func (this *Sample) StringTst1(connObj *rpc.RpcConnection) string {
	return "你好"
}

func (this *Sample) StringTst2(connObj *rpc.RpcConnection, name string) string {
	return "你好:" + name
}

// 多参数多返回值
func (this *Sample) StringTst3(connObj *rpc.RpcConnection, name string, name2 string) (string, string) {
	return "你好1：" + name, "你好2:" + name2
}

type Man struct {
	Name string
	Sex  int
}

func (this *Sample) StructTst1(connObj *rpc.RpcConnection, man Man) (result Man) {
	man.Name = "Server" + man.Name
	man.Sex = 10 + man.Sex

	return man
}
