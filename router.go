package fido

import (
	"context"
	"net/http"
	"regexp"
)

type Router struct {
	routes     []Route
	middleware []Middleware
	handler    http.Handler
}

type Route struct {
	method  string
	path    string
	matcher *regexp.Regexp
	handler http.Handler
}

func NewRouter() *Router {
	r := &Router{routes: make([]Route, 0)}
	r.handler = r.newHandler()

	return r
}

// Handle associates a handler to a route path. Route paths may contain regexp
// to extract named path components. e.g. the following path could be used to
// fetch a person by numeric id: "/people/(?P<id>\\d+)"
// Paths are tested in order, the first registered path that matches is used.
// Paths match against the entire path, each path regexp is surrounded with "^"
// and "$" before testing the regexp.
// The supplied handler must be of type http.Handler or
// func(http.ResponseWriter, *http.Request)
func (r *Router) Handle(method string, path string, handler any) {
	var h http.Handler
	switch v := handler.(type) {
	case http.Handler:
		h = v
	case func(http.ResponseWriter, *http.Request):
		h = http.HandlerFunc(v)
	default:
		panic("Unknown handler type")
	}
	route := Route{method, path, regexp.MustCompile("^" + path + "$"), h}
	r.routes = append(r.routes, route)
}

// Use registers the supplied middleware
func (r *Router) Use(mw Middleware) {
	r.middleware = append(r.middleware, mw)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	next := r.handler

	for i := len(r.middleware) - 1; i >= 0; i-- {
		next = r.middleware[i](next)
	}

	next.ServeHTTP(w, req)
}

func (r *Router) newHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.routes {
			if route.method != req.Method || !route.matcher.MatchString(req.URL.Path) {
				continue
			}
			ctx := context.WithValue(
				req.Context(), ParamsKey, buildPathParams(route, req),
			)
			route.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		JSONError(w, http.StatusNotFound, "resource could not be found")
	}
}
