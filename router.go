package fido

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

// Router is a http.Handler capable of routing requests.
type Router struct {
	routes     []route
	middleware []Middleware
	handler    http.Handler
}

type route struct {
	method  string
	path    string
	matcher *regexp.Regexp
	handler http.Handler
}

// NewRouter returns a new router instance.
func NewRouter() *Router {
	r := &Router{routes: make([]route, 0)}
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
	rt := route{method, path, regexp.MustCompile("^" + path + "$"), h}
	r.routes = append(r.routes, rt)
}

// Use registers the supplied middleware. Middleware are run in the order they
// are added.
func (r *Router) Use(mw Middleware) {
	r.middleware = append(r.middleware, mw)
}

// ServeHTTP calls all registered middleware, then routes the request to the
// registered handler or returns 404 if a matching handler cannot be found.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	next := r.handler

	for i := len(r.middleware) - 1; i >= 0; i-- {
		next = r.middleware[i](next)
	}

	next.ServeHTTP(w, req)
}

func (r *Router) newHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, rt := range r.routes {
			if rt.method != req.Method || !rt.matcher.MatchString(req.URL.Path) {
				continue
			}
			ctx := context.WithValue(
				req.Context(), ParamsKey, buildPathParams(rt, req),
			)
			rt.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		JSONError(w, http.StatusNotFound, "resource could not be found")
	}
}

// Path parameter extraction

type paramsKey struct{}

// ParamsKey can be used to fetch the path params from the request's context.
// Typically the Params function is used directly.
var ParamsKey = paramsKey{}

type pathParams struct {
	params map[string]string
}

// PathParams defines an interface for accessing path parameters by string or
// by int.
type PathParams interface {
	Get(string) string
	GetInt(string) (int, error)
}

// Get returns the named path parameter as a string.
func (p *pathParams) Get(key string) string {
	if v, ok := p.params[key]; ok {
		return v
	}
	return ""
}

// GetInt returns the named path parameter as an integer.
func (p *pathParams) GetInt(key string) (int, error) {
	v, ok := p.params[key]
	if !ok {
		return 0, fmt.Errorf("value not found")
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func buildPathParams(rt route, req *http.Request) PathParams {
	params := make(map[string]string)
	names := rt.matcher.SubexpNames()
	for i, match := range rt.matcher.FindStringSubmatch(req.URL.Path) {
		if names[i] != "" {
			params[names[i]] = match
		}
	}

	return &pathParams{params}
}

// Params extracts the path paremeters from the requests context and returns an
// object which implements PathParams
func Params(req *http.Request) PathParams {
	return req.Context().Value(ParamsKey).(PathParams)
}
