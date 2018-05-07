//	Copyright 2014 willsky. All rights reserved.
//	Use of this source code is governed by a GNU
//  license that can be found in the LICENSE file.
//
//	Author email: hdu_willsky@foxmail.com
//	Version: 1.0
//
//  Introduction：
//	SlimGo log implementator. provide file log function.
//

package slimgo

import (
	"errors"
	"fmt"
	"os"
)

type fileLog struct {
	file *os.File
}

// Register this log.
func init() {
	Register("file", newFileLog())
}

// new File log
func newFileLog() ILogger {
	f := new(fileLog)
	return f
}

// init File log
func (f *fileLog) Init(configStr string) error {
	var config struct {
		FileName string `json:"filename"`
	}

	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return err
	}

	if len(config.FileName) == 0 {
		return errors.New("filename can not be null")
	}

	// open log file and set it to default output.
	if file, err := OpenORCreateFile(config.FileName); err != nil {
		return err
	} else {
		f.file = file
	}

	return nil
}

// imp Message method
func (f *fileLog) Message(message string, level LogLevel) error {
	_, err := fmt.Fprintln(f.file, message)
	return err
}

// imp Flush method
func (f *fileLog) Flush() {
	f.file.Sync()
}

// imp Close method
func (f *fileLog) Close() {
	f.Flush()
	f.file.Close()
}
