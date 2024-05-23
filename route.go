package grouter

import "net/http"

// Route hold info about HTTP_METHOD, PATH, handler func and a list of middlewares
type Route struct {
	method      string
	path        string
	handler     http.Handler
	middlewares []func(http.Handler) http.Handler
}

// With adds a middleware to a route
func (r *Route) With(middlewares ...func(http.Handler) http.Handler) *Route {
	r.middlewares = middlewares
	return r
}

func (r *Route) Do(handler http.Handler) *Route {
	r.handler = handler
	return r
}

func (r *Route) DoFunc(handler http.HandlerFunc) *Route {
	r.handler = handler
	return r
}
