package grouter

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// router is for syntactic sugary
// It allows for more readable http.Server mux routes as in router.HTTP_METHOD(PATH).Do(HANDLER).with(MIDDLEWARE)
// At the end we get a http.ServeMux

// Router holds all middlewares and routes. it is used to generate a http.ServeMux
type Router struct {
	pathPrefix  string
	middlewares []func(http.Handler) http.Handler
	routes      []*Route
	muxOnce     sync.Once
	mux         *http.ServeMux
}

// NewRouter returns a new Router struct. you can add middlewares directly to this Router that will act globally (like logging, tracing)
func NewRouter(pathPrefix string, middlewares ...func(http.Handler) http.Handler) *Router {
	return &Router{
		pathPrefix:  pathPrefix,
		middlewares: middlewares,
	}
}

// joinPaths is a helper function to add a prefixPath (url path) to all routes
func (r *Router) joinPaths(paths ...string) string {
	combinedPath := make([]string, 0)
	if r.pathPrefix != "/" {
		combinedPath = append(combinedPath, r.pathPrefix)
	}
	combinedPath = append(combinedPath, paths...)
	return strings.Join(combinedPath, "/")
}

// Any adds a route for ALL http methods to the Router and returns a route
func (r *Router) Any(path string) *Route {
	_route := &Route{method: "", path: path}
	r.routes = append(r.routes, _route)
	return _route
}

// Get adds a route for GET http methods to the Router and returns a route
func (r *Router) Get(path string) *Route {
	_route := &Route{method: http.MethodGet, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

// Post adds a route for POST http methods to the Router and returns a route
func (r *Router) Post(path string) *Route {
	_route := &Route{method: http.MethodPost, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

func (r *Router) Delete(path string) *Route {
	_route := &Route{method: http.MethodDelete, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

func (r *Router) Put(path string) *Route {
	_route := &Route{method: http.MethodPut, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

func (r *Router) Patch(path string) *Route {
	_route := &Route{method: http.MethodPatch, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

func (r *Router) Head(path string) *Route {
	_route := &Route{method: http.MethodHead, path: path}
	r.routes = append(r.routes, _route)
	return _route
}

// GetMux returns (and generates) the http.ServerMux from the routes in the Router
func (r *Router) GetMux() *http.ServeMux {
	r.muxOnce.Do(func() {
		r.mux = http.NewServeMux()
		for _, rr := range r.routes {
			if rr.handler == nil {
				continue
			}
			// the handler from that route
			h := rr.handler

			// add all middlewares from the Router
			for _, mw := range r.middlewares {
				h = mw(h)
			}

			// add all middlewares from the route
			for _, mw := range rr.middlewares {
				h = mw(h)
			}

			// then add this all into our mux and let golang handle it
			r.mux.Handle(strings.Trim(fmt.Sprintf("%s %s", rr.method, r.joinPaths(rr.path)), " "), h)
		}
	})

	return r.mux
}
