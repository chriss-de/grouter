# grouter [![PkgGoDev](https://pkg.go.dev/badge/github.com/chriss-de/grouter)](https://pkg.go.dev/github.com/chriss-de/grouter) [![Go Report Card](https://goreportcard.com/badge/github.com/chriss-de/grouter)](https://goreportcard.com/report/github.com/chriss-de/grouter)

A little router that helps with net/http mux. It makes the code clearer and handles middlewares for http route.
It is intended to be used with golang 1.23+ and net/http server mux.

# Usage

```go
    import "github.com/chriss-de/grouter/v2"

    // base_url of your API/webapp ist https://domain/url
    router := grouter.NewRouter("/url")

    // this adds a route for any http verb for /api/docs/
    router.Any("/api/docs/").DoFunc(httpSwagger.Handler(...))
	
    // this adds a healthz endpoint
    router.Get("/healthz").DoFunc(func(writer http.ResponseWriter, request *http.Request) {
        writer.WriteHeader(http.StatusOK)
    })

    // REST API
    restV1.RegisterEndpoints(router.AddSubRouter("/api/v1"))
	
// -- 
func RegisterEndpoints(r *grouter.Router) {
    r.AddMiddlewares(setContentType)

    r.Get("/objects").DoFunc(handlers.GetObjects)
    r.Get("/objects/{ID}").DoFunc(handlers.GetObjectByID)
}
```

This will result in following routes in the mux

| url path                 | http method | call                                                                      |
|--------------------------|-------------|---------------------------------------------------------------------------|
| /url/api/docs/           | ANY         | An http handler function                                                  |
| /url/healthz             | GET         | will always return HTTP/200/OK                                            |
| /url/api/v1/objects      | GET         | will call handlers.GetObjects                                             |
| /url/api/v1/objects/{ID} | GET         | will call handlers.GetObjectByID where the path variable `ID` can be used |

# Middlewares

You can add any middleware. You can add them to the router or to any route.

```go
    import "github.com/chriss-de/grouter/v2"

    // base_url of your API/webapp ist https://domain/url
    router := grouter.NewRouter("/url", middlewares.MW1, middlewares.MW2)
	
    router.Get("/healthz").With(middleware.MW3, middleware.MW4).Do(handler)
    router.Get("/stuff").With(middleware.MW5).Do(stuff)

```

| url path     | call chain through middlewares  |
|--------------|---------------------------------|
| /url/healthz | MW1 > MW2 > MW3 > MW4 > handler |
| /url/stuff   | MW1 > MW2 > MW5 > stuff         |
