package utils

import (
	"testing"
	"reflect"
	"path"
	"strings"
	"fmt"
)

func TestStringToBytes(t *testing.T) {
	s := "111xfa哇"
	bytes := StringToBytes(&s)
	str := BytesToString(&bytes)
	if s != str {
		t.Error("bytes != string")
	}
}

type e struct {
	sss string
}

func TestGetType(t *testing.T) {
	var i interface{}

	typeName := reflect.ValueOf(i).Type().String()
	t.Log(typeName)
}

func TestGetFuncInfo(t *testing.T) {
	fileName, file, line := GetFuncInfo(TestGetType)
	fileName = fileName[ strings.LastIndexByte(fileName, '.')+1:]
	_, file = path.Split(file)
	s := fmt.Sprintf("[%s %d] %s", file, line, fileName)
	t.Log(s)
}
