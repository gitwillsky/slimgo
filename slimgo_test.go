package slimgo

import (
	"fmt"
	"testing"
	"time"
)

func Test_Server(t *testing.T) {
	s := New()
	s.Use(func(c Context) {
		start := time.Now()
		c.Next()
		fmt.Printf("use time : %.3f s\n", time.Since(start).Seconds())
	})

	s.GET("/post/:id", func(c Context) {
		id := c.Param("id")
		c.String(200, id)
	})

	s.Root("/api", func(c Context) {
		c.Next()
	}).GET("/user/:id", func(c Context) {
		id := c.Param("id")
		time.Sleep(time.Second * 3)
		c.String(200, id)
	})

	if err := s.Start(":8080"); err != nil {
		t.Error(err)
	}
}
