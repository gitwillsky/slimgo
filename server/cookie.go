// cookie unit
package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
