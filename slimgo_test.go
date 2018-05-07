package slimgo

import (
	"fmt"
	"testing"
	"time"
)

func Test_Server(t *testing.T) {
	s := New()
	s.AddServerFilter(func(c *Context) (interface{}, error) {
		start := time.Now()
		r, e := c.Next()
		fmt.Printf("use time : %.3f s\n", time.Since(start).Seconds())
		return r, e
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
