package fido

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type H map[string]any

func JSON(w http.ResponseWriter, status int, d any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Error writing JSON")
	}
}

func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, H{"error": message})
}

func BindJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// Path parameter extraction

type paramsKey struct{}

var ParamsKey = paramsKey{}

type pathParams struct {
	params map[string]string
}

type PathParams interface {
	Get(string) string
	GetInt(string) (int, error)
}

// Get returns the named path parameter as a string
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

func buildPathParams(route Route, req *http.Request) PathParams {
	params := make(map[string]string)
	names := route.matcher.SubexpNames()
	for i, match := range route.matcher.FindStringSubmatch(req.URL.Path) {
		if names[i] != "" {
			params[names[i]] = match
		}
	}

	return &pathParams{params}
}

func Params(req *http.Request) PathParams {
	return req.Context().Value(ParamsKey).(PathParams)
}
