package grouter

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// router is for syntactic sugary
// It allows for more readable http.Server serveMux routes as in router.HTTP_METHOD(PATH).Do(HANDLER).with(MIDDLEWARE)
// At the end we get a http.ServeMux

// Router holds all middlewares and routes. it is used to generate a http.ServeMux
type Router struct {
	serveMux    http.Handler
	pathPrefix  string
	middlewares []func(http.Handler) http.Handler
	subRouters  []*Router
	routes      []*Route
	onceLock    sync.Once
}

// NewRouter returns a new Router struct. you can add middlewares directly to this Router that will act globally (like logging, tracing)
func NewRouter(pathPrefix string, middlewares ...func(http.Handler) http.Handler) *Router {
	if pathPrefix == "" {
		pathPrefix = "/"
	}

	return &Router{
		pathPrefix:  pathPrefix,
		middlewares: middlewares,
		routes:      make([]*Route, 0),
		subRouters:  make([]*Router, 0),
	}
}

// joinPaths is a helper function to add a prefixPath (url path) to all routes
func (r *Router) renderPath(paths ...string) string {
	combinedPath := make([]string, 0)
	if r.pathPrefix == "/" {
		combinedPath = append(combinedPath, "")
	} else {
		combinedPath = append(combinedPath, r.pathPrefix)
	}
	if len(paths) == 1 && paths[0] == "" {
		paths = nil
	}
	combinedPath = append(combinedPath, paths...)
	return strings.Join(combinedPath, "/")
}

func (r *Router) AddMiddlewares(middlewares ...func(http.Handler) http.Handler) *Router {
	for _, mw := range middlewares {
		r.middlewares = append(r.middlewares, mw)
	}
	return r
}

func (r *Router) AddSubRouter(path string) *Router {
	sr := NewRouter(r.renderPath(strings.TrimPrefix(path, "/")))
	r.subRouters = append(r.subRouters, sr)
	return sr
}

func (r *Router) AddRoute(path string, methods ...string) *Route {
	route := newRoute(r.renderPath(strings.TrimPrefix(path, "/")), methods...)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Any(path string) *Route {
	return r.AddRoute(path, "")
}

func (r *Router) GetHead(path string) *Route {
	return r.AddRoute(path, http.MethodGet, http.MethodHead)
}

// Get adds a route for GET http methods to the Router and returns a route
func (r *Router) Get(path string) *Route {
	return r.AddRoute(path, http.MethodGet)
}

// Post adds a route for POST http methods to the Router and returns a route
func (r *Router) Post(path string) *Route {
	return r.AddRoute(path, http.MethodPost)
}

func (r *Router) Delete(path string) *Route {
	return r.AddRoute(path, http.MethodDelete)
}

func (r *Router) Put(path string) *Route {
	return r.AddRoute(path, http.MethodPut)
}

func (r *Router) Patch(path string) *Route {
	return r.AddRoute(path, http.MethodPatch)
}

func (r *Router) Head(path string) *Route {
	return r.AddRoute(path, http.MethodHead)
}

func (r *Router) getServeMux(serveMux *http.ServeMux) {
	// generate subRouters for router
	for _, subRouter := range r.subRouters {
		if subRouter == nil {
			continue
		}
		subRouter.getServeMux(serveMux)
	}

	// generate route for router
	for _, route := range r.routes {
		if route.handler == nil {
			continue
		}
		// the handler from that route
		routeHandler := route.handler

		// add all middlewares from the route
		for _, routeMiddleware := range route.middlewares {
			routeHandler = routeMiddleware(routeHandler)
		}

		// then add this all into our serveMux and let golang handle it
		for _, routeMethod := range route.method {
			serveMux.Handle(strings.Trim(fmt.Sprintf("%s %s", routeMethod, route.path), " "), routeHandler)
		}
	}

	//
	r.serveMux = serveMux

	// add all middlewares from the Router
	for _, routerMiddleware := range r.middlewares {
		r.serveMux = routerMiddleware(r.serveMux)
	}
}

// GetServeMux returns (and generates) the http.ServerMux from the routes in the Router
func (r *Router) GetServeMux() http.Handler {
	r.onceLock.Do(func() {
		serveMux := http.NewServeMux()
		r.getServeMux(serveMux)
	})

	return r.serveMux
}
