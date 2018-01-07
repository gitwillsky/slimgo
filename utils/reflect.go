package utils

import (
	"runtime"
	"reflect"
)

// GetFuncInfo 通过反射获取方法信息
func GetFuncInfo(i interface{}) (funcName, file string, line int) {
	pc := reflect.ValueOf(i).Pointer()
	funcInfo := runtime.FuncForPC(pc)
	funcName = funcInfo.Name()
	file, line = funcInfo.FileLine(pc)
	return
}
