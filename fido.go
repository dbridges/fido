// Package fido is a minimalist collection of utility functions for making HTTP
// JSON APIs.
//
// fido requires Go 1.18 or newer
//
// See https://github.com/dbridges/fido for in-depth examples.

package fido

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// H is an alias for map[string]any, useful for quickly generating JSON objects
type H map[string]any

// JSON encodes a value and writes it to the supplied response writer after
// setting the response status.
func JSON(w http.ResponseWriter, status int, d any) {
	err := json.NewEncoder(w).Encode(d)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Error writing JSON")
	}
}

// JSONError wraps the message in a JSON object and writes it to the supplied
// response writer
func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, H{"error": message})
}

// BindJSON decodes the http request body into v
func BindJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
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
