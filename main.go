package gout

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// RouterHook 测试中间件
func RouterHook() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		// 计算请求解析时间
		c.Next()

		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

type request struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type response struct {
	Code    int         `json:"code" xml:"code"`
	Message string      `json:"message" xml:"message"`
	Data    interface{} `json:"data" xml:"data"`
}

func main() {
	//r := gout.New(gout.DefaultOption)
	r := NewServer()

	r.Use(
		Recovery(),
		//gzip.Gzip(gzip.DefaultCompression),
		//RouterHook(),
	)

	r.GET("/index", func(c *Context) {
		names := []string{"test_str"}
		c.String(http.StatusOK, names[100])
	})

	v1 := r.Group("/v1")
	{
		v1.GET("", func(c *Context) {
			c.Html(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *Context) {
			msg := fmt.Sprintf("hello %s, you're at %s", c.Param("name"), c.Path)
			c.JSON(http.StatusOK, H{"code": 200, "message": msg})
		})
		v2.GET("/hello2", func(c *Context) {
			var resp = &response{
				200,
				"test message,test message",
				"sdgsdg",
			}
			c.JSON(http.StatusOK, resp)
		})

		v2.GET("/xml", func(c *Context) {
			var resp = &response{
				200,
				"test message",
				"sdgsdg",
			}
			c.Xml(http.StatusOK, resp)
		})

		v2.GET("/r", func(c *Context) {
			c.Redirect(http.StatusMovedPermanently, "https://www.baidu.com")
		})

		v2.POST("/login", func(c *Context) {
			c.JSON(http.StatusOK, H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

		v2.POST("/test", func(c *Context) {
			var req = &request{}
			d := c.Req.Header.Get("aaa")
			err := c.JsonParse(req)
			if err != nil {
				panic(err)
			}
			c.SetHeader("Server", "gout server")
			c.JSON(http.StatusOK, H{
				"Id":   req.Id,
				"Name": req.Name,
				"DDD":  d,
			})
		})

		v2.POST("/file", func(c *Context) {
			file, err := c.FormFile("file")

			if err != nil {
				c.Fail(400, err.Error())
				return
			}

			err = c.SaveUploadedFile(file, fmt.Sprintf("upload/%s", file.Filename))
			if err != nil {
				c.Fail(400, err.Error())
				return
			}

			c.JSON(200, H{
				"file": file.Filename,
				"size": file.Size,
			})
		})
	}

	r.Run(":7055")
}
