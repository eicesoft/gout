package gout

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
)

const (
	_PayloadName = "_payload"
)

const (
	ContextTypeHeaderName = "Content-Type"
)

const (
	MIMEJSON  = "application/json"
	MIMEHTML  = "text/html"
	MIMEXML   = "application/xml"
	MIMEPlain = "text/plain"
)

const abortIndex int8 = math.MaxInt8 / 2

// H 用于返回JSON数据
type H map[string]interface{}
type ParamMap map[string]string
type KeyMap map[string]interface{}

// Context 存储请求上下文信息
type Context struct {
	// 其他对象
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	Params ParamMap
	// 响应信息
	StatusCode int
	// 中间件
	handlers HandlersChain
	index    int8
	Engine   *Engine

	mu   sync.RWMutex
	Keys KeyMap
}

// NewContext 构建上下文实例
func NewContext(w http.ResponseWriter, r *http.Request, engine *Engine) *Context {
	return &Context{
		Req:        r,
		Writer:     w,
		StatusCode: 200,
		Path:       r.URL.Path,
		Method:     r.Method,
		Engine:     engine,
		index:      -1, // 用于记录执行到那个中间件
	}
}

func (c *Context) reset() {
	c.handlers = nil
	c.index = -1
	c.Keys = nil
}

func (c *Context) init(w http.ResponseWriter, req *http.Request, handlers HandlersChain) {
	c.Req = req
	c.Writer = w
	c.Path = req.URL.Path
	c.Method = req.Method
	c.Keys = map[string]interface{}{}
	c.handlers = handlers
}

// Next 所属
func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Fail 直接中断响应
func (c *Context) Fail(code int, err string) {
	c.index = abortIndex
	c.JSON(code, H{"message": err, "code": http.StatusInternalServerError})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Payload(payload interface{}) {
	c.Set(_PayloadName, payload)
}

func (c *Context) getPayload() interface{} {
	if payload, ok := c.Get(_PayloadName); ok {
		return payload
	}

	return nil
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}

	c.Keys[key] = value
	c.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	c.mu.RUnlock()
	return
}

func (c *Context) JsonParse(obj interface{}) error {
	decoder := json.NewDecoder(c.Req.Body)

	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Req.MultipartForm == nil {
		if err := c.Req.ParseMultipartForm(c.Engine.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}

	f, fh, err := c.Req.FormFile(name)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return fh, err
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			c.Fail(http.StatusBadRequest, err.Error())
		}
	}(src)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			c.Fail(http.StatusBadRequest, err.Error())
		}
	}(out)

	_, err = io.Copy(out, src) //Copy file
	return err
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
	//c.Writer.WriteHeader(code)
}

// SetHeader 设置header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) WriteBody(code int, buf interface{}, contextType string) {
	if contextType != "" {
		c.SetHeader(ContextTypeHeaderName, contextType)
	}
	c.Status(code)
	c.Payload(buf)
}

// 快速构建响应
// 返回字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.WriteBody(code, fmt.Sprintf(format, values...), MIMEPlain)
}

// JSON 返回json数据
func (c *Context) JSON(code int, obj interface{}) {
	c.WriteBody(code, obj, MIMEJSON)
}

// Data 返回字节流数据
func (c *Context) Data(code int, data []byte) {
	c.WriteBody(code, data, "")
}

// HTML 返回html数据
func (c *Context) HTML(code int, html string) {
	c.WriteBody(code, html, MIMEHTML)
}
