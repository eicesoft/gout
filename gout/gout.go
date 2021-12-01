package gout

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type HandlerFunc func(c *Context)

// Engine 作为最顶层
type Engine struct {
	http.Handler //实现Handler
	server       *http.Server
	*RouterGroup // 具备单个路由的GET POST方法
	router       *router
	groups       []*RouterGroup // 存储所有的分组
}

// RouterGroup 管理各种路由
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // 支持路由分组中间件
	parent      *RouterGroup  // support nesting
	engine      *Engine       // 所有的路由分组 共享一个 engine
}

// New 新建一个 实例
func New() *Engine {
	engine := &Engine{router: newRouter()}

	// RouterGroup里面的 engine属性为 自身的engine  确保所有的engine 为一个
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc

	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := NewContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)
}

func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}

// Group is defined to create a new RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 方法直接放在分组路由上
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 方法同上
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// PUT PUT路由
func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	log.Printf("Listen in address %s", addr)
	engine.server = &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 28, //256M
	}

	return engine.server.ListenAndServe()
}
