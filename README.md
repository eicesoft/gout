## Gout

A simple like gin framework, this name is gout ^&^
包含路由， 错误处理， 中间件等基本功能。 也实现了一个简单的gzip压缩的中间件作为参考。

这个是一个Demo

```go
r := gout.New(&gout.Option{
    IsEnablePProf: true,
})
r.Use(
    gout.Recovery(),
    //gzip.Gzip(gzip.DefaultCompression),
    //RouterHook(),
)

//router group
v2 := r.Group("/v2")
{
    v2.GET("/hello/:name", func(c *gout.Context) {
        msg := fmt.Sprintf("hello %s, you're at %s", c.Param("name"), c.Path)
        c.JSON(http.StatusOK, gout.H{"code": 200, "message": msg})
    })

    v2.GET("/hello2", func(c *gout.Context) {
        var resp = &response{
            200,
            "test message",
            "sdgsdg",
        }
        c.JSON(http.StatusOK, resp)
    })

    //xml response
    v2.GET("/xml", func(c *gout.Context) {
        var resp = &response{
            200,
            "test message",
            "sdgsdg",
        }
        c.Xml(http.StatusOK, resp)
    })

    //redirect response
    v2.GET("/r", func(c *gout.Context) {
        c.Redirect(http.StatusMovedPermanently, "https://www.baidu.com")
    })

    //json response
    v2.POST("/test", func(c *gout.Context) {
        var req = &request{}
        d := c.Req.Header.Get("aaa")
        err := c.JsonParse(req)
        if err != nil {
            panic(err)
        }
        c.SetHeader("Server", "gout server")
        c.JSON(http.StatusOK, gout.H{
            "Id":   req.Id,
            "Name": req.Name,
            "DDD":  d,
        })
    })

    //upload
    v2.POST("/file", func(c *gout.Context) {
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

        c.JSON(200, gout.H{
            "file": file.Filename,
            "size": file.Size,
        })
    })
}

r.Run(":7055")
```

这个框架基本上只是实现了一个Web框架最基础的部分. 但麻雀虽小, 五脏俱全. 一些简单的项目还是可以用的. 编译出的文件也比较的小. 适合写一些小型项目. 