package main

import (
	"fmt"
	"github.com/eicesoft/web_server/gout"
	"log"
	"net/http"
	"time"
)

func RouterHook() gout.HandlerFunc {
	return func(c *gout.Context) {
		// Start timer
		t := time.Now()

		// 如果错误 可以直接返回
		//c.Fail(500, "Internal Server Error")

		// 计算请求解析时间
		c.Next()
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := gout.New()

	// 捕获所有异常中间件
	r.Use(gout.Recovery())
	// 使用自定义中间件
	r.Use(RouterHook())

	r.GET("/index", func(c *gout.Context) {
		names := []string{"test_str"}
		c.String(http.StatusOK, names[100])
	})

	v1 := r.Group("/v1")
	{
		v1.GET("", func(c *gout.Context) {
			time.Sleep(500)
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *gout.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *gout.Context) {
			msg := fmt.Sprintf("hello %s, you're at %s", c.Param("name"), c.Path)
			c.JSON(http.StatusOK, gout.H{"code": 200, "message": msg})
		})
		v2.POST("/login", func(c *gout.Context) {
			c.JSON(http.StatusOK, gout.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	_ = r.Run("127.0.0.1:7055")
}