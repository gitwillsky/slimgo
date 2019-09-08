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

	type form struct {
		Name string `form:"name"`
	}

	s.GET("/post/:id/:name", func(c Context) {
		var f form
		if err := c.Bind(&f); err != nil {
			c.String(400, err.Error())
			return
		}
		c.String(200, fmt.Sprintf("%s %s %s", f.Name, c.Param("id"), c.Param("name")))
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
