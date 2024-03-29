package gout

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HandlerFunc defines the handler used by middleware.
type HandlerFunc func(c *Context)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

const defaultMultipartMemory = 32 << 21 // 64 MB

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
	PUT    = "PUT"
)

var DefaultOption = &Options{
	false,
}

// Engine 作为最顶层
type Engine struct {
	http.Handler       //实现Handler
	*RouterGroup       // 具备单个路由的GET POST方法
	server             *http.Server
	router             *router
	groups             []*RouterGroup // 存储所有的分组
	pool               sync.Pool
	MaxMultipartMemory int64 //MaxMultipartMemory
}

// RouterGroup 管理各种路由
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // 支持路由分组中间件
	parent      *RouterGroup  // support nesting
	engine      *Engine       // 所有的路由分组 共享一个 engine
}

// NewServer 新建一个 实例
func NewServer(opts ...Option) *Engine {
	ui := `
	██████╗   ██████╗ ██╗   ██╗████████╗
	██╔════╝ ██╔═══██╗██║   ██║╚══██╔══╝
	██║  ███╗██║   ██║██║   ██║   ██║
	██║   ██║██║   ██║██║   ██║   ██║
	╚██████╔╝╚██████╔╝╚██████╔╝   ██║
	╚═════╝  ╚═════╝  ╚═════╝    ╚═╝`
	fmt.Println(ui)
	engine := &Engine{
		router:             newRouter(),
		MaxMultipartMemory: defaultMultipartMemory,
	}

	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}

	// RouterGroup里面的 engine属性为 自身的engine  确保所有的engine 为一个
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}

	options := newOptions(opts...)
	if options.IsEnablePProf {
		log.Printf("* Registry pprof routers - /debug/pprof")
		WrapPProfHandler(engine)
	}

	return engine
}

func (engine *Engine) allocateContext() *Context {
	return &Context{Engine: engine, index: -1, StatusCode: 200, value: &Values{}}
}

func (engine *Engine) handleRequest(c *Context) {
	var middlewares HandlersChain

	for _, group := range engine.groups {
		if strings.HasPrefix(c.Req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c.handlers = middlewares
	engine.router.handle(c)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.reset()
	c.init(w, req)
	engine.handleRequest(c)

	engine.pool.Put(c)
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
	//log.Printf("Route add [%4s] - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 方法直接放在分组路由上
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodGet, pattern, handler)
}

// POST 方法同上
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPost, pattern, handler)
}

// PUT PUT路由
func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPut, pattern, handler)
}

// Run Start a http server
func (engine *Engine) Run(addr string) {
	log.Printf("Listen in address %s", addr)
	engine.server = &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 28, //256M
	}

	go func() {
		if err := engine.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server startup err %s", err.Error())
		}
	}()

	ExitHook().Close(func() { //处理信号进行关闭
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := engine.server.Shutdown(ctx); err != nil {
			panic(err)
		} else {
			log.Printf("Server is closed.")
		}
	})
}
