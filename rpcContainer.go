package rpc

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"

	"github.com/polariseye/rpc-go/log"
)

type RpcContainer struct {
	funcData         map[string]*MethodInfo
	getConvertorFunc func() IByteConvertor

	// 请求超时时间
	requestExpireMillisecond int64

	// 数据的字节序
	byteOrder binary.ByteOrder
}

// 注册一个RPC服务端
func (this *RpcContainer) RegisterService(obj interface{}) {
	tp := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	// 提取所有公有函数
	clsName := this.getClassName(tp)

	for i := 0; i < tp.NumMethod(); i++ {
		mthd := tp.Method(i)
		mthdVal := val.Method(i)

		err := this.addRpcMethod(clsName, mthd.Name, mthd.Type, mthdVal, true)
		if err != nil {
			panic(err)
		}
	}
}

func (this *RpcContainer) getClassName(tp reflect.Type) string {
	for {
		if tp.Kind() == reflect.Struct {
			break
		}

		tp = tp.Elem()
	}

	return tp.Name()
}

func (this *RpcContainer) RegisterFunc(moduleName string, methodName string, funcObj interface{}) {
	tp := reflect.TypeOf(funcObj)
	val := reflect.ValueOf(funcObj)

	err := this.addRpcMethod(moduleName, methodName, tp, val, false)
	if err != nil {
		panic(err)
	}
}

func (this *RpcContainer) addRpcMethod(moduleName string, methodName string, methodType reflect.Type, methodVal reflect.Value, isFromStruct bool) error {
	// 获取参数
	paramList := make([]reflect.Type, 0, methodType.NumIn())
	for i := 0; i < methodType.NumIn(); i++ {
		if isFromStruct && i == 0 {
			// 因为struct的第一个输入参数是struct自身，而调用时，不需要传这个参数，所以跳过
			continue
		}

		paramList = append(paramList, methodType.In(i))
	}

	// 获取返回
	returnList := make([]reflect.Type, 0, methodType.NumOut())
	for i := 0; i < methodType.NumOut(); i++ {
		returnList = append(returnList, methodType.Out(i))
	}

	// 第一个必须是连接对象类型的
	if len(paramList) <= 0 || paramList[0] != RpcConnectionType {
		return fmt.Errorf("Param invalid ModuleName:%s MethodName:%s", moduleName, methodName)
	}

	/*
		// 返回值最后一个必须是error
		if len(returnList) <= 0 || returnList[len(returnList)-1] != ErrorType {
			return errors.New("Return invalid ")
		}
	*/

	name := fmt.Sprintf("%s_%s", moduleName, methodName)
	if _, exist := this.funcData[name]; exist {
		return fmt.Errorf("rpc repeated:%s", name)
	}

	mthdInfoItem := newMethodInfo(name, methodVal, paramList, returnList)
	this.funcData[name] = mthdInfoItem

	return nil
}

// 获取一个连接对象
func (this *RpcContainer) GetRpcConnection(con net.Conn) *RpcConnection {
	return NewRpcConnection(this, con)
}

func (this *RpcContainer) getMethod(methodName string) (*MethodInfo, bool) {
	result, exist := this.funcData[methodName]

	return result, exist
}

func (this *RpcContainer) RecordAllMethod() {
	for methodName, item := range this.funcData {
		log.Debug("MethodName:%v ParamCount:%v ReturnCount:%v", methodName, len(item.funcParamList), len(item.returnValueList))
	}
}

func NewRpcContainer(getConvertorFunc func() IByteConvertor) *RpcContainer {
	return &RpcContainer{
		funcData:                 make(map[string]*MethodInfo, 8),
		getConvertorFunc:         getConvertorFunc,
		byteOrder:                binary.BigEndian,
		requestExpireMillisecond: 2 * 60 * 1000,
	}
}
