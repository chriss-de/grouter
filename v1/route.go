package grouter

import "net/http"

// Route hold info about HTTP_METHOD, PATH, handler func and a list of middlewares
type Route struct {
	method      []string
	path        string
	handler     http.Handler
	middlewares []func(http.Handler) http.Handler
}

func newRoute(path string, methods ...string) *Route {
	return &Route{method: methods, path: path}
}

// With adds a middleware to a route
func (r *Route) With(middlewares ...func(http.Handler) http.Handler) *Route {
	r.middlewares = middlewares
	return r
}

func (r *Route) Do(handler http.Handler) {
	r.handler = handler
}

func (r *Route) DoFunc(handler http.HandlerFunc) {
	r.handler = handler
}
