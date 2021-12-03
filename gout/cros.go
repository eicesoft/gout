package gout

import "net/http"

type Cors struct {
	allowedOrigins   []string
	allowedHeaders   []string
	allowedMethods   []string
	allowCredentials bool
}

func CROSHook() HandlerFunc {
	return func(c *Context) {

	}
}

func (c *Cors) Handler(h http.Handler) http.Handler {
	return nil
}
