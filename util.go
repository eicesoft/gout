package gout

import "net/http"

// H 用于返回JSON数据
type H map[string]interface{}

// WrapF is a helper function for wrapping http.HandlerFunc and returns a middleware.
func WrapF(f http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		f(c.Writer, c.Req)
	}
}

// WrapH is a helper function for wrapping http.Handler and returns a middleware.
func WrapH(h http.Handler) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.Writer, c.Req)
	}
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
