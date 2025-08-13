package grouter

import (
	"net/http"
	"strings"
	"sync"
)

// router is for syntactic sugary
// It allows for more readable http.Server serveMux routes as in router.HTTP_METHOD(PATH).With(MIDDLEWARE).Do(HANDLER)
// At the end we get an http.ServeMux

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
	sr.middlewares = append(sr.middlewares, r.middlewares...)
	r.subRouters = append(r.subRouters, sr)
	return sr
}

// AddRoute adds a route to a router. A path and a list of HTTP methods/verbs
func (r *Router) AddRoute(path string, methods ...string) *Route {
	route := newRoute(r.renderPath(strings.TrimPrefix(path, "/")), methods...)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Any(path string) *Route     { return r.AddRoute(path, "") }
func (r *Router) Get(path string) *Route     { return r.AddRoute(path, http.MethodGet) }
func (r *Router) Post(path string) *Route    { return r.AddRoute(path, http.MethodPost) }
func (r *Router) Delete(path string) *Route  { return r.AddRoute(path, http.MethodDelete) }
func (r *Router) Put(path string) *Route     { return r.AddRoute(path, http.MethodPut) }
func (r *Router) Patch(path string) *Route   { return r.AddRoute(path, http.MethodPatch) }
func (r *Router) Head(path string) *Route    { return r.AddRoute(path, http.MethodHead) }
func (r *Router) Options(path string) *Route { return r.AddRoute(path, http.MethodOptions) }

func (r *Router) GetHead(path string) *Route {
	return r.AddRoute(path, http.MethodGet, http.MethodHead)
}

func (r *Router) generateMux(serveMux *http.ServeMux) {
	// generate route for router
	for _, route := range r.routes {
		if route.handler == nil {
			continue
		}
		// the handler from that route
		routeHandler := route.handler

		// add all middlewares from the route
		for idx := len(route.middlewares) - 1; idx >= 0; idx-- {
			routeHandler = route.middlewares[idx](routeHandler)
		}
		// add all middlewares from the Router
		for idx := len(r.middlewares) - 1; idx >= 0; idx-- {
			routeHandler = r.middlewares[idx](routeHandler)
		}
		// then add this all into our serveMux and let golang handle it
		for _, routeMethod := range route.methods {
			serveMux.Handle(strings.Trim(routeMethod+" "+route.path, " "), routeHandler)
		}
	}
	// generate subRouters for router
	for _, subRouter := range r.subRouters {
		if subRouter == nil {
			continue
		}
		subRouter.generateMux(serveMux)
	}
}

// ServeHTTP to implement the http.Handler interface
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.onceLock.Do(func() {
		serveMux := http.NewServeMux()
		r.generateMux(serveMux)
		r.serveMux = serveMux
	})
	r.serveMux.ServeHTTP(rw, req)
}

// GetServeMux returns (and generates) the http.ServerMux from the routes in the Router
func (r *Router) GetServeMux() http.Handler {
	r.onceLock.Do(func() {
		serveMux := http.NewServeMux()
		r.generateMux(serveMux)
		r.serveMux = serveMux
	})

	return r.serveMux
}
