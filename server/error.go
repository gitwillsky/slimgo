package server

import (
	"runtime"
	"fmt"
	"time"
	"github.com/gitwillsky/slimgo/utils"
)

type handlerError struct {
	Date       string   `json:"time"`
	StatusCode int      `json:"code"`
	Messages   []string `json:"messages"`
	Debug      string   `json:"debug"`
}

func (c *Context) NewError(statusCode int, errs ...error) error {
	result := &handlerError{
		StatusCode: statusCode,
	}
	for _, e := range errs {
		result.Messages = append(result.Messages, e.Error())
	}

	result.Date = time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(1)
	result.Debug = fmt.Sprintf("%s [%d]", file, line)

	return result
}

func (e *handlerError) Error() string {
	data, _ := json.Marshal(e)
	return utils.BytesToString(&data)
}
