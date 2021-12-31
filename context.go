package gout

import (
	"encoding/json"
	"fmt"
	"github.com/eicesoft/gout/render"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
)

const abortIndex int8 = math.MaxInt8 / 2
const defaultMemory = 32 << 20

const (
	MIMEJson              = "application/json"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

// ParamMap 参数类型
type ParamMap map[string]string

// dataMap 上下文参数
type dataMap map[string]interface{}

// Context 存储请求上下文信息
type Context struct {
	writermem  responseWriter
	index      int8           //中间件执行索引
	value      *Values        //上下文参数
	handlers   HandlersChain  //中间件数组
	Writer     ResponseWriter // Writer 响应接口
	Req        *http.Request  // Req http 请求结构
	Path       string         // Path 请求路径
	Method     string         // Method 请求方法
	Params     ParamMap       // Params 请求参数
	StatusCode int            //响应状态码
	Engine     *Engine        //服务器引擎
}

// NewContext 构建上下文实例
//func NewContext(w http.ResponseWriter, r *http.Request, engine *Engine) *Context {
//	return &Context{
//		Req:        r,
//		Writer:     w,
//		StatusCode: 200,
//		Path:       r.URL.Path,
//		Method:     r.Method,
//		Engine:     engine,
//		index:      -1, // 用于记录执行到那个中间件
//	}
//}

// Values defined
type Values struct {
	value dataMap      //上下文参数
	mu    sync.RWMutex //dataMap 同步锁
}

// IValue Values结构接口实现
type IValue interface {
	reset()                                          // reset Values重置
	Set(key string, value interface{})               // Set 设置hash Key值
	Get(key string) (value interface{}, exists bool) // Get 获得 Value 对应 key 值
}

// reset Values重置
func (d *Values) reset() {
	d.value = nil
}

// Set 设置hash Key值
func (d *Values) Set(key string, value interface{}) {
	d.mu.Lock()
	if d.value == nil {
		d.value = make(map[string]interface{})
	}

	d.value[key] = value
	d.mu.Unlock()
}

// Get 获得 Value 对应 key 值
func (d *Values) Get(key string) (value interface{}, exists bool) {
	d.mu.RLock()
	value, exists = d.value[key]
	d.mu.RUnlock()
	return
} // Values End

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.handlers = nil
	c.index = -1
	c.value.reset()
	c.Path = ""
}

func (c *Context) init(w http.ResponseWriter, req *http.Request) {
	c.Req = req
	c.writermem.reset(w)
	c.Path = req.URL.Path
	c.Method = req.Method
}

// Next 所属
func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
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
}

// SetHeader 设置header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

// Render write response body buffer
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)
	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}

// String 返回字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, render.Text{Data: fmt.Sprintf(format, values...)})
}

// JSON 返回Json数据
func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, render.JSON{Data: obj})
}

// Raw 返回字节流数据
func (c *Context) Raw(code int, data []byte) {
	c.Render(code, render.Raw{Data: data})
}

// Html 返回html数据
func (c *Context) Html(code int, html string) {
	c.Render(code, render.HTML{Data: html})
}

func (c *Context) Template(code int, filename string, values interface{}) {
	c.Render(code, render.Template{Data: values, Filename: filename})
}

func (c *Context) Xml(code int, xml interface{}) {
	c.Render(code, render.XML{Data: xml})
}

func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{Code: code, Request: c.Req, Location: location})
}

func (c *Context) Success(data interface{}) {
	c.JSON(http.StatusOK, &H{
		"code":    200,
		"data":    data,
		"message": "",
	})
}

// Fail 直接中断响应
func (c *Context) Fail(code int, err string) {
	c.index = abortIndex
	c.JSON(http.StatusInternalServerError, H{"message": err, "code": code})
}

func (c *Context) shouldBindWith(obj interface{}, b Binding) error {
	return b.Bind(c.Req, obj)
}

func (c *Context) mustBindWith(obj interface{}, b Binding) error {
	if err := c.shouldBindWith(obj, b); err != nil {
		return err
	}

	return nil
}

func (c *Context) GetHeader(key string) string {
	return c.Req.Header.Get(key)
}

func (c *Context) ContentType() string {
	return filterFlags(c.GetHeader("Content-Type"))
}

func (c *Context) Bind(obj interface{}) error {
	b := Default(c.Req.Method, c.ContentType())
	return c.mustBindWith(obj, b)
}

func Default(method, contentType string) Binding {
	if method == http.MethodGet {
		return QueryBind
	} else {
		switch contentType {
		case MIMEJson:
			return JsonBind
		case MIMEMultipartPOSTForm:
			return FormMultipartBind
		default:
			return FormBind
		}
	}
}

func (c *Context) BindForm(obj interface{}) error {
	return c.mustBindWith(obj, FormBind)
}

func (c *Context) BindJson(obj interface{}) error {
	return c.mustBindWith(obj, JsonBind)
}

func (c *Context) BindQuery(obj interface{}) error {
	return c.mustBindWith(obj, QueryBind)
}

func (c *Context) BindFormPost(obj interface{}) error {
	return c.mustBindWith(obj, FormPostBind)
}

func (c *Context) BindFormMultipart(obj interface{}) error {
	return c.mustBindWith(obj, FormMultipartBind)
}
