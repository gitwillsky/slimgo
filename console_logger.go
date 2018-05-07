//	Copyright 2014 willsky. All rights reserved.
//	Use of this source code is governed by a GNU
//  license that can be found in the LICENSE file.
//
//	Author email: hdu_willsky@foxmail.com
//	Version: 1.0
//
//  Introduction：
//	SlimGo log implementator. provide console log function.
//

package slimgo

import (
	"fmt"
	"io"
	"os"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;37"), // Emergency     	white
	newBrush("1;36"), // Alert			    cyan
	newBrush("1;35"), // Critical         magenta
	newBrush("1;31"), // Error            red
	newBrush("1;33"), // Warning          yellow
	newBrush("1;32"), // Notice			green
	newBrush("1;34"), // Informational 	blue
	newBrush("1;34"), // Debug            blue
}

type consoleLog struct {
	writerCloser io.WriteCloser
}

// new console log
func newConsole() ILogger {
	c := new(consoleLog)
	return c
}

// init console log
func (c *consoleLog) Init(config string) error {
	c.writerCloser = os.Stdout
	return nil
}

// imp Message method
func (c *consoleLog) Message(message string, level LogLevel) error {
	_, err := fmt.Fprintln(c.writerCloser, colors[level](message))
	return err
}

// imp Flush method
func (c *consoleLog) Flush() {}

// imp Close method
func (c *consoleLog) Close() {
	//c.writerCloser.Close()
}
