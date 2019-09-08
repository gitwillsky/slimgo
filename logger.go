//	Copyright 2014 willsky. All rights reserved.
//	Use of this source code is governed by a GNU
//  license that can be found in the LICENSE file.
//
//	Author email: hdu_willsky@foxmail.com
//	Version: 1.0
//
package slimgo

import "log"

type Logger interface {
	Debugf(format string, fields ...interface{})
	Infof(format string, fields ...interface{})
	Errorf(format string, fields ...interface{})
}

type logger struct {
	debug bool
}

func (l logger) Debugf(format string, fields ...interface{}) {
	if l.debug {
		log.Printf("[DEBUG] "+format, fields...)
	}
}

func (logger) Infof(format string, fields ...interface{}) {
	log.Printf("[INFO] "+format, fields...)
}

func (logger) Errorf(format string, fields ...interface{}) {
	log.Printf("[ERROR] "+format, fields...)
}
