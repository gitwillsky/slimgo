package server

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gitwillsky/slimgo/log"
	"github.com/gitwillsky/slimgo/utils"
	"fmt"
	"path"
)

// param router param
type param struct {
	key   string
	value string
}

type Handler func(context *Context) (interface{}, error)

type params []param

// Context server context
type Context struct {
	regPath  string // 注册handler时的地址
	response http.ResponseWriter
	Request  *http.Request
	data     *sync.Map
	params   params // 路由参数
	handlers []Handler
	index    int16
	wrote    bool
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// newContext create a context
func newContext(res http.ResponseWriter, req *http.Request, handlers []Handler) *Context {
	c := contextPool.Get().(*Context)
	c.response = res
	c.Request = req
	c.data = &sync.Map{}
	c.handlers = make([]Handler, len(handlers), 10)
	copy(c.handlers, handlers)
	c.index = -1
	return c
}

// release release context
func (c *Context) release() {
	c.regPath = ""
	c.response = nil
	c.Request = nil
	c.data = nil
	c.params = nil
	c.handlers = nil
	c.wrote = false // important
	contextPool.Put(c)
}

// GetValue get data value
func (c *Context) GetValue(key string) (value interface{}, ok bool) {
	return c.data.Load(key)
}

func (c *Context) PutValue(key string, value interface{}) {
	c.data.Store(key, value)
}

// GetParam get router :paramname param value
func (c *Context) GetParam(key string) string {
	for _, v := range c.params {
		if v.key == key {
			return v.value
		}
	}
	return ""
}

func (c *Context) GetRegURLPath() string {
	return c.regPath
}

func (c *Context) AddHandlers(handlers ...Handler) {
	c.handlers = append(c.handlers, handlers...)
}

func (c *Context) GetAllHandlers() []Handler {
	return c.handlers
}

func (c *Context) GetResponseHeader() http.Header {
	return c.response.Header()
}

func (c *Context) WriteHeader(code int) {
	c.response.WriteHeader(code)
	c.wrote = true
}

func (c *Context) Next() (interface{}, error) {
	c.index++
	for l := int16(len(c.handlers)); c.index < l; c.index++ {
		handler := c.handlers[c.index]
		if log.GetLevel() == log.LevelDebug {
			fileName, file, line := utils.GetFuncInfo(handler)
			fileName = fileName[ strings.LastIndexByte(fileName, '.')+1:]
			_, file = path.Split(file)
			if c.index < l-1 {
				log.Debugf("execute filter {%s} for request: [%s] %s", fmt.Sprintf("%s[%d]:%s", file, line, fileName),
					c.Request.Method, c.Request.URL.Path)
			} else {
				log.Debugf("execute handler {%s} for request: [%s] %s", fmt.Sprintf("%s[%d]:%s", file, line, fileName),
					c.Request.Method, c.Request.URL.Path)
			}
		}
		if result, err := handler(c); result != nil || err != nil {
			// Abort now!
			return result, err
		}
	}
	return nil, nil
}
