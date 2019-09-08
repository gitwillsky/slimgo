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
	"regexp"
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

type Handler func(context Context)

type params []param

type Context interface {
	Data(key string) (value interface{}, ok bool)
	PutData(key string, value interface{})
	Param(key string) string
	RegRelativePath() string
	Request() *http.Request
	ResponseWriter() ResponseWriter
	ResponseHeader() http.Header
	WriteResponseHeader(code int)
	JSON(statusCode int, data interface{})
	String(statusCode int, data string)
	BindJSON(target interface{}, validators ...func(target interface{}) error) error
	SaveFiles(folder string, maxLen int, allowExt string) ([]string, error)
	SetCookie(key string, value string, cookiePath string, maxAge int) error
	SetSecureCookie(secret, cookieName, cookieValue, cookiePath string, cookieMaxDay int) error
	Cookie(key string) string
	SecureCookie(secret, key string) string
	ClientIP() string
	Next()
	Abort()
}

// context server context
type context struct {
	regPath  string // 注册handler时的地址
	response ResponseWriter
	request  *http.Request
	data     *sync.Map
	params   *params // 路由参数
	handlers []Handler
	index    int
	server   *Server
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &context{}
	},
}

// newContext create a context
func newContext() *context {
	c := contextPool.Get().(*context)
	return c
}

func (c *context) init(s *Server, res http.ResponseWriter, req *http.Request, middleware ...Handler) {
	c.server = s
	c.response = newResponseWriter(res)
	c.request = req
	c.data = &sync.Map{}
	c.handlers = append(c.handlers, middleware...)

	// find router handler
	handlers, regPath, ps := c.server.router.GetHandlers(req)
	c.regPath = regPath
	c.params = ps
	c.handlers = append(c.handlers, handlers...)
}

// release release context
func (c *context) recycle() {
	c.server.router.psRecycle(c.params)

	c.regPath = ""
	c.response = nil
	c.request = nil
	c.data = nil
	c.params = nil
	c.handlers = c.handlers[:0]
	c.index = 0
	contextPool.Put(c)
}

// Data get data value
func (c *context) Data(key string) (value interface{}, ok bool) {
	return c.data.Load(key)
}

func (c *context) PutData(key string, value interface{}) {
	c.data.Store(key, value)
}

// Param get router :paramname param value
func (c *context) Param(key string) string {
	if c.params != nil {
		for _, v := range *c.params {
			if v.key == key {
				return v.value
			}
		}
	}

	return ""
}

func (c *context) RegRelativePath() string {
	return c.regPath
}

func (c *context) ResponseWriter() ResponseWriter {
	return c.response
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) ResponseHeader() http.Header {
	return c.response.Header()
}

func (c *context) WriteResponseHeader(code int) {
	c.response.WriteHeader(code)
}

func (c *context) run() {
	for l := len(c.handlers); c.index < l; c.index++ {
		handler := c.handlers[c.index]

		if c.server.mode == Debug {
			fileName, file, line := GetFuncInfo(handler)
			fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
			_, file = path.Split(file)
			c.server.logger.Debugf("%s %s {%s}",
				c.request.Method, c.request.URL.Path,
				fmt.Sprintf("%s[%d]:%s", file, line, fileName))
		}

		handler(c)
		if c.response.Written() {
			return
		}
	}
}

func (c *context) Next() {
	c.index++
	c.run()
}

func (c *context) Abort() {
	c.index = len(c.handlers)
}

func (c *context) JSON(statusCode int, data interface{}) {
	c.response.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.WriteResponseHeader(statusCode)

	if err := json.NewEncoder(c.response).Encode(data); err != nil {
		http.Error(c.response, err.Error(), 500)
		return
	}
}

func (c *context) String(statusCode int, data string) {
	c.response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.response.Header().Set("X-Content-Type-Options", "nosniff")
	c.response.WriteHeader(statusCode)
	_, _ = fmt.Fprintln(c.response, data)
}

// 文件接收
// floder 文件所要保存的目录
// maxLen 文件最大长度
// 允许的扩展名[正则表达式]，例如：(.png|.jpeg|.jpg|.gif)
func (c *context) SaveFiles(folder string, maxLen int, allowExt string) ([]string, error) {
	result := make([]string, 0)
	// 允许的扩展名正则匹配
	regExt, err := regexp.Compile(allowExt)
	if err != nil {
		return nil, err
	}

	// 文件长度限制
	length, err := strconv.Atoi(c.request.Header.Get("Content-Length"))
	if err != nil || (length > maxLen) {
		return nil, errors.New("Upload file is too big!")
	}

	// get the multipart reader for the request.
	reader, err := c.request.MultipartReader()
	if err != nil {
		return nil, err
	}

	// 文件全路径
	var fullFilePath string
	// 读取每个文件,保存。
	for {
		// 通过boundary，获取每个文件数据。
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
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
		_ = os.MkdirAll(path.Dir(fullFilePath), os.ModePerm)
		// 建立目标文件
		dst, err := os.Create(fullFilePath)
		if err != nil {
			return nil, err
		}
		defer dst.Close()

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

// BindJSON parse json request
func (c *context) BindJSON(target interface{}, validators ...func(target interface{}) error) error {
	contentType := c.request.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return errors.New("invalid Content-Type : " + contentType)
	}

	dec := json.NewDecoder(c.request.Body)
	defer c.request.Body.Close()

	if err := dec.Decode(target); err != nil {
		return err
	}

	for _, validator := range validators {
		if err := validator(target); err != nil {
			return err
		}
	}

	return nil
}

// Proxy returns proxy client ips slice.
func proxy(r *http.Request) string {
	// standard format: X-ForWard-For: client, proxy1, proxy2..
	if ips, ok := r.Header["X-Forwarded-For"]; ok {
		return ips[0]
	}
	return ""
}

// ClientIP() return client IP.
// if in proxy, return first proxy id;
// if error ,return 127.0.0.1;
func (c *context) ClientIP() string {
	clientIP := strings.TrimSpace(c.request.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = c.request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.request.RemoteAddr)); err == nil {
		return ip
	}
	return "127.0.0.1"
}

// Set cookie.
func (c *context) SetCookie(key string, value string, cookiePath string, maxAge int) error {
	var b bytes.Buffer

	_, _ = fmt.Fprintf(&b, "%s=%s", strings.TrimSpace(key), strings.TrimSpace(value))

	// set max age.
	if maxAge >= 0 {
		_, _ = fmt.Fprintf(&b, "; Max-Age=%d", maxAge)
	}
	// set path.
	_, _ = fmt.Fprintf(&b, "; Path=%s", strings.TrimSpace(cookiePath))

	// output to header.
	c.response.Header().Add("Set-Cookie", b.String())

	return nil
}

// Set secure cookie.
func (c *context) SetSecureCookie(secret, cookieName, cookieValue, cookiePath string, cookieMaxDay int) error {
	// encoding value to string.
	s := base64.URLEncoding.EncodeToString([]byte(cookieValue))
	// get time stamp.
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	// new hmac to crypto secret.
	h := hmac.New(sha1.New, []byte(secret))
	_, _ = fmt.Fprintf(h, "%s%s", s, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	// v|timestamp|sig
	cookie := strings.Join([]string{s, timestamp, sig}, "|")

	_ = c.SetCookie(cookieName, cookie, cookiePath, cookieMaxDay)

	return nil
}

// Get secure cookie.
func (c *context) SecureCookie(secret, key string) string {
	cookie := c.Cookie(key)

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
	_, _ = fmt.Fprintf(h, "%s%s", s, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return ""
	}

	res, _ := base64.URLEncoding.DecodeString(s)
	return string(res)
}

// Get cookie
func (c *context) Cookie(key string) string {
	cookie, err := c.request.Cookie(key)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}
