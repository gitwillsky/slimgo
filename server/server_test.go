package server

import (
	"fmt"
	"testing"
	"time"
)

func Test_Server(t *testing.T) {
	s := New()
	s.AddFilter(func(c *Context) (interface{}, error) {
		start := time.Now()
		c.Next()
		fmt.Printf("use time : %vs\n", time.Since(start).Seconds())
		return nil, nil
	})

	s.GET("/post/:id", func(c *Context) (interface{}, error) {
		id := c.GetParam("id")
		c.WriteHeader(200)
		return id, nil
	})

	s.Root("/api", func(c *Context) (interface{}, error) {
		c.Next()
		return nil, nil
	}).
		GET("/user/:id", func(c *Context) (interface{}, error) {
		id := c.GetParam("id")
		time.Sleep(time.Second * 3)
		c.WriteHeader(200)
		return id, nil
	})

	s.Start(":8080")
}
