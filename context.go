package slimgo

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
	response ResponseWriter
	Request  *http.Request
	data     *sync.Map
	params   params // 路由参数
	handlers []Handler
	index    int
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// newContext create a context
func newContext(res http.ResponseWriter, req *http.Request, handlers []Handler) *Context {
	c := contextPool.Get().(*Context)
	c.response = NewResponseWriter(res)
	c.Request = req
	c.data = &sync.Map{}
	c.handlers = make([]Handler, len(handlers)*2)
	copy(c.handlers, handlers)
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
	c.index = 0
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

func (c *Context) GetResponseWriter() http.ResponseWriter {
	return c.response
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
}

func (c *Context) Next() (interface{}, error) {
	for l := len(c.handlers); c.index < l; c.index++ {
		handler := c.handlers[c.index]

		// log
		if GetLevel() == LevelDebug {
			fileName, file, line := GetFuncInfo(handler)
			fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
			_, file = path.Split(file)
			if c.index < l-1 {
				Debugf("execute filter {%s} for request: [%s] %s", fmt.Sprintf("%s[%d]:%s", file, line, fileName),
					c.Request.Method, c.Request.URL.Path)
			} else {
				Debugf("execute handler {%s} for request: [%s] %s", fmt.Sprintf("%s[%d]:%s", file, line, fileName),
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

func (c *Context) resolveHandlerResult(data interface{}, e error) {
	if data == nil && e == nil {
		return
	}

	if e != nil {
		code := 500
		contentType := "text/plain; charset=utf-8"
		if err, ok := e.(*handlerError); ok {
			contentType = "application/json;charset=utf-8"
			code = err.StatusCode
		}
		c.response.Header().Set("Content-Type", contentType)
		c.response.WriteHeader(code)
		fmt.Fprintf(c.response, e.Error())
		return
	}

	if data != nil {
		reflectValue := reflect.ValueOf(data)
	WALK:
		typeKind := reflectValue.Type().Kind().String()
		typeName := reflectValue.Type().String()
		switch typeKind {
		case "string":
			fmt.Fprint(c.response, data)
		case "ptr":
			reflectValue = reflectValue.Elem()
			goto WALK
		case "struct":
			// file
			if typeName == "os.File" {
				file := data.(*os.File)
				d, err := file.Stat()
				if err != nil {
					http.Error(c.response, err.Error(), 500)
					return
				}
				http.ServeContent(c.response, c.Request, file.Name(), d.ModTime(), file)
				return
			}
			// convert struct to json and output response
			writeJson(c.response, data)
		case "map", "array", "slice":
			writeJson(c.response, data)
		case "int":
			if !c.response.Written() {
				c.WriteHeader(data.(int))
				return
			}
		default:
			http.Error(c.response, "No resource resolver for "+typeName+", please add resouce filter to handle this", 500)
		}
	}
}

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
	if GetLevel() == LevelDebug {
		_, file, line, _ := runtime.Caller(1)
		result.Debug = fmt.Sprintf("%s [%d]", file, line)
	}

	return result
}

func (e *handlerError) Error() string {
	data, _ := json.Marshal(e)
	return BytesToString(&data)
}

func writeJson(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json;charset=utf-8")
	dat, err := json.Marshal(data)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	fmt.Fprint(res, BytesToString(&dat))
}

// 文件上传
// floder 文件所要保存的目录
// maxLen 文件最大长度
// 允许的扩展名[正则表达式]，例如：(.png|.jpeg|.jpg|.gif)
func (c *Context) UploadFiles(folder string, maxLen int, allowExt string) ([]string, error) {
	result := make([]string, 0)
	// 允许的扩展名正则匹配
	regExt, err := regexp.Compile(allowExt)
	if err != nil {
		return nil, err
	}

	// 文件长度限制
	length, err := strconv.Atoi(c.Request.Header.Get("Content-Length"))
	if err != nil || (length > maxLen) {
		return nil, errors.New("Upload file is too big!")
	}

	// get the multipart reader for the request.
	reader, err := c.Request.MultipartReader()
	if err != nil {
		return nil, err
	}

	// 文件全路径
	var fullFilePath string
	// 读取每个文件,保存。
	for {
		// 通过boundary，获取每个文件数据。
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		// 检查文件名
		fileName := part.FileName()
		if fileName == "" {
			continue
		}
		// 检查文件类型
		fileExt := strings.ToLower(path.Ext(fileName))
		if !regExt.MatchString(fileExt) {
			continue
			//return nil, errors.New("Invalid upload file type!")
		}
		// 构造新的文件名和全路径
		subFilePath := fmt.Sprintf("%s/%s", time.Now().Format("2006-01-02"), fileName)
		fullFilePath = path.Clean(folder + subFilePath)

		// 新建文件夹,0777
		os.MkdirAll(path.Dir(fullFilePath), os.ModePerm)
		// 建立目标文件
		dst, err := os.Create(fullFilePath)
		defer dst.Close()
		if err != nil {
			return nil, err
		}

		// 拷贝数据流到文件
		if _, err = io.Copy(dst, part); err != nil {
			return nil, err
		}

		// 将成功写入的文件路径信息写入结果集
		result = append(result, subFilePath)
	}

	// 检查结果
	if len(result) == 0 {
		return nil, errors.New("Invalid upload file!")
	}

	return result, nil
}

// ParseJSONRequest parse json request
func (c *Context) ParseJSONRequest(target interface{}) error {
	if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
		dec := json.NewDecoder(c.Request.Body)
		err := dec.Decode(target)
		c.Request.Body.Close()
		return err
	}

	return errors.New("Content-Type isn't application/json")
}

// Proxy returns proxy client ips slice.
func proxy(r *http.Request) string {
	// standard format: X-ForWard-For: client, proxy1, proxy2..
	if ips, ok := r.Header["X-Forwarded-For"]; ok {
		return ips[0]
	}
	return ""
}

// GetClientIP return client IP.
// if in proxy, return first proxy id;
// if error ,return 127.0.0.1;
func (c *Context) GetClientIP() string {
	clientIP := strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = c.Request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return "127.0.0.1"
}

// Set cookie.
func (c *Context) SetCookie(key string, value string, cookiePath string, maxAge int) error {
	var b bytes.Buffer

	fmt.Fprintf(&b, "%s=%s", strings.TrimSpace(key), strings.TrimSpace(value))

	// set max age.
	if maxAge >= 0 {
		fmt.Fprintf(&b, "; Max-Age=%d", maxAge)
	}
	// set path.
	fmt.Fprintf(&b, "; Path=%s", strings.TrimSpace(cookiePath))

	// output to header.
	c.response.Header().Add("Set-Cookie", b.String())

	return nil
}

// Set secure cookie.
func (c *Context) SetSecureCookie(secret, cookieName, cookieValue,
cookiePath string, cookieMaxDay int) error {

	// encoding value to string.
	s := base64.URLEncoding.EncodeToString([]byte(cookieValue))
	// get time stamp.
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	// new hmac to crypto secret.
	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", s, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	// v|timestamp|sig
	cookie := strings.Join([]string{s, timestamp, sig}, "|")

	c.SetCookie(cookieName, cookie, cookiePath, cookieMaxDay)

	return nil
}

// Get secure cookie.
func (c *Context) GetSecureCookie(secret, key string) string {
	cookie := c.GetCookie(key)

	if cookie == "" {
		return ""
	}

	parts := strings.SplitN(cookie, "|", 3)

	if len(parts) != 3 {
		return ""
	}

	s := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	// judge secret correct.
	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", s, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return ""
	}

	res, _ := base64.URLEncoding.DecodeString(s)
	return string(res)
}

// Get cookie
func (c *Context) GetCookie(key string) string {
	cookie, err := c.Request.Cookie(key)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}
