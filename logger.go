//	Copyright 2014 willsky. All rights reserved.
//	Use of this source code is governed by a GNU
//  license that can be found in the LICENSE file.
//
//	Author email: hdu_willsky@foxmail.com
//	Version: 1.0
//
//  Introduction：
//	SlimGo App log package. provide App log function.
//
//	Usage:
//
//	import "github.com/SlimGo/core/log"
//
//	log := log.New(10)
//
//	log.Alert("Alert info: %s",alert) 			// Alert
//	log.Error("error")	//error

package slimgo

import (
	"bytes"
)

type LogLevel uint8

// RFC5424 log message levels.
const (
	LevelEmergency LogLevel = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// Logger interface
type ILogger interface {
	Init(config string) error
	Message(message string, level LogLevel) error
	Flush()
	Close()
}

var implements = make(map[string]ILogger, 2)

// Register log interface implementator
func Register(name string, impl ILogger) {
	if impl == nil {
		panic("log implement Object is nil")
	}
	if _, ok := implements[name]; ok {
		panic("log implement Object exists")
	}

	implements[name] = impl
}

func makeFormats(a []interface{}) string {
	buf := &bytes.Buffer{}
	for i := 0; i < len(a); i++ {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString("%v")
	}
	return buf.String()
}
