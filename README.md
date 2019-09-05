# SlimGo Web Framework

![slimgo logo](./logo.png)

又一个 go web 框架，wheel。

# 起步

#### 安装

```bash
go get github.com/gitwillsky/slimgo
```

#### 使用

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/gitwillsky/slimgo"
)

func main() {
	s := slimgo.New()

	// 全局过滤器 global filter
	// 全局过滤器在route之前执行
    s.AddServerFilter(func(ctx *server.Context) (interface{}, error) {
        now := time.Now()
		r, e := ctx.Next()

		// 这里可以定义router handler的结果处理，比如自定义json解析逻辑， 模板渲染逻辑，自定义错误处理等

		// ctx.getRegURLPath() 方法获得注册路由时的URL原始字符串，global filter 由系统定义为 "/*"
        log.Debugf("Handler %s in %f seconds", ctx.GetRegURLPath(), time.Since(now).Seconds())
        return r, e
    })

	// 注册单个handler
	s.GET("/hello", func(ctx *server.Context) (interface{}, error) {
		return "hello world", nil
	})

	//  根（组）路由支持
	s.Root("/system", func(ctx *server.Context) (interface{}, error){
		log.Infof("进入根路由")
		// 这里适合进行路由权限校验等逻辑
		// 这里如果返回结果或者错误，那么下面的handler将不会执行
		return nil, nil
	}).
		GET("/files/*filepath", system.StaticFileHandler).

		POST("/token", system.TokenHandler).

		// 也可以单独为后面的handler做其他的权限校验
		// 这里如果不调用ClearFilters() 那么根上定义的filter也将应用到后面的handler
		ClearFilters().
		AddFilter(filter.AuthFilter).

		POST("/files", system.UploadFileHandler).

		GET("/my", system.GetMyInfoHandler).

		PATCH("/my", system.PatchMyInfoHandler).

		PUT("/mypass", system.UpdatePasswordHandler)

    // 启动web服务
    if err = s.Start(":8080"); err != nil {
        log.Errorf("start web server failed, %s", err.Error())
    }
}
```

# 参考

1.  beego (github.com/astaxie/beego)
1.  http-router(github.com/julienschmidt/httprouter)

# 联系

Author Email: hdu_willsky@foxmail.com
