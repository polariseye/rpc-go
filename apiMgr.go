package rpc

import (
	"fmt"
	"reflect"

	"github.com/polariseye/rpc-go/log"
)

type ApiMgr struct {
	funcData map[string]*MethodInfo
}

// 注册一个RPC服务端
func (this *ApiMgr) RegisterService(obj interface{}) {
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

func (this *ApiMgr) getClassName(tp reflect.Type) string {
	for {
		if tp.Kind() == reflect.Struct {
			break
		}

		tp = tp.Elem()
	}

	return tp.Name()
}

func (this *ApiMgr) RegisterFunc(moduleName string, methodName string, funcObj interface{}) {
	tp := reflect.TypeOf(funcObj)
	val := reflect.ValueOf(funcObj)

	err := this.addRpcMethod(moduleName, methodName, tp, val, false)
	if err != nil {
		panic(err)
	}
}

func (this *ApiMgr) addRpcMethod(moduleName string, methodName string, methodType reflect.Type, methodVal reflect.Value, isFromStruct bool) error {
	// 获取参数
	paramList := make([]reflect.Type, 0, methodType.NumIn())
	for i := 0; i < methodType.NumIn(); i++ {
		if isFromStruct && i == 0 {
			// 因为struct的第一个输入参数是struct自身，而调用时，不需要传这个参数，所以跳过
			continue
		}

		tpItem := methodType.In(i)
		paramList = append(paramList, tpItem)
	}

	// 获取返回
	returnList := make([]reflect.Type, 0, methodType.NumOut())
	for i := 0; i < methodType.NumOut(); i++ {
		returnList = append(returnList, methodType.Out(i))
	}

	// 第一个必须是连接对象类型的
	if len(paramList) <= 0 || paramList[0] != RpcConnectionerType {
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

func (this *ApiMgr) getMethod(methodName string) (*MethodInfo, bool) {
	result, exist := this.funcData[methodName]

	return result, exist
}

func (this *ApiMgr) RecordAllMethod() {
	for methodName, item := range this.funcData {
		log.Debug("MethodName:%v ParamCount:%v ReturnCount:%v", methodName, len(item.funcParamList), len(item.returnValueList))
	}
}

func newApiMgr() *ApiMgr {
	return &ApiMgr{
		funcData: make(map[string]*MethodInfo, 8),
	}
}
