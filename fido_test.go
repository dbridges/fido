package fido

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	handleTestJSON := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			JSON(w, http.StatusOK, H{"message": "test"})
		})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handleTestJSON.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "{\"message\":\"test\"}\n")
	contentType := res.Header.Get("Content-Type")
	assert(t, contentType, "application/json; charset=utf-8")
}

func TestJSONError(t *testing.T) {
	handleTestJSON := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			JSONError(w, http.StatusBadRequest, "bad request")
		})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handleTestJSON.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusBadRequest)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "{\"error\":\"bad request\"}\n")
	contentType := res.Header.Get("Content-Type")
	assert(t, contentType, "application/json; charset=utf-8")
}

func TestBindJSON(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}
	r := httptest.NewRequest(
		http.MethodGet,
		"/",
		strings.NewReader("{\"name\": \"test user\"}"),
	)
	var itm item
	err := BindJSON(r, &itm)
	assert(t, err, nil)
	assert(t, itm.Name, "test user")
}
