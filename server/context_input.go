package server

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

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