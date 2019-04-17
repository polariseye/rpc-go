package rpc

import "unsafe"

// 判断一个数据是否为nil。包含把一个具体类型赋值给interface的情况
// 返回值:
// bool: true：为nil false:不为nil
func IsNil(val interface{}) bool {
	if val == nil {
		return true
	}

	type InterfaceStructure struct {
		pt uintptr // 指向interface方法表的指针
		pv uintptr // 指向对应值的指针
	}
	is := *(*InterfaceStructure)(unsafe.Pointer(&val))
	if is.pv == 0 {
		return true
	}

	return false
}
