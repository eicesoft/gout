package gout

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 用于返回JSON数据
type H map[string]interface{}

// Context 存储请求上下文信息
type Context struct {
	// 其他对象
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	Params map[string]string
	// 响应信息
	StatusCode int
	// 中间件
	handlers []HandlerFunc
	index    int
}

// NewContext 构建上下文实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Req:        r,
		Writer:     w,
		StatusCode: 200,
		Path:       r.URL.Path,
		Method:     r.Method,
		index:      -1, // 用于记录执行到那个中间件
	}
}

// Next 所属
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Fail 直接中断响应
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)

	c.JSON(code, H{"message": err, "code": http.StatusInternalServerError})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// Query 获取url的查询参数
func (c *Context) Query(name string) string {
	return c.Req.URL.Query().Get(name)
}

// PostForm 获取表单参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// Status 设置状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader 设置header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 快速构建响应
// 返回字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain;charset=utf-8")
	c.Status(code)
	_, _ = c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 返回json数据
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json;charset=utf-8")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Data 返回字节流数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, _ = c.Writer.Write(data)
}

// HTML 返回html数据
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html;charset=utf-8")
	c.Status(code)
	_, _ = c.Writer.Write([]byte(html))
}
