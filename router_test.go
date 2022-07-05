package fido

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"testing"
)

func assert(t *testing.T, got any, expected any) {
	if got == expected {
		return
	}
	msg := fmt.Sprintf("Expected '%v', but got '%v'", expected, got)
	if _, file, line, ok := runtime.Caller(1); ok {
		t.Errorf("\n%s:%d: %s", path.Base(file), line, msg)
	} else {
		t.Error(msg)
	}
}

func TestRouterBasicMatching(t *testing.T) {
	handleGetItem := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get item"))
	}
	handlePostItem := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("post item"))
	}
	handleGetOtherItem := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get other item"))
	}

	router := NewRouter()
	router.Handle("GET", "/item", handleGetItem)
	router.Handle("POST", "/item", handlePostItem)
	router.Handle("GET", "/item/other", handleGetOtherItem)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/item", nil)
	router.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "get item")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/item", nil)
	router.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "post item")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/item/other", nil)
	router.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "get other item")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/item/not/exists", nil)
	router.ServeHTTP(w, r)
	assert(t, w.Result().StatusCode, http.StatusNotFound)
}

func TestRouterRegexpMatching(t *testing.T) {
	handleGetInt := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get int"))
	}
	handleGetString := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get string"))
	}

	router := NewRouter()
	router.Handle("GET", "/item/(?P<id>[0-9]+)", handleGetInt)
	router.Handle("GET", "/item/(?P<id>[a-z]+)", handleGetString)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/item/42", nil)
	router.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "get int")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/item/thing", nil)
	router.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "get string")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/item/BAD", nil)
	router.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusNotFound)
}

func TestRouterRegexpCapture(t *testing.T) {
	handleGetById := func(w http.ResponseWriter, r *http.Request) {
		id := Params(r).Get("id")
		w.Write([]byte("id is " + id))
	}

	handleGetByIdInt := func(w http.ResponseWriter, r *http.Request) {
		id, err := Params(r).GetInt("id")
		if err != nil {
			w.Write([]byte("something went wrong"))
			return
		}
		w.Write([]byte(fmt.Sprintf("id is %d", id)))
	}

	router := NewRouter()
	router.Handle("GET", "/item/(?P<id>[a-z]+)", handleGetById)
	router.Handle("GET", "/item/(?P<id>[0-9]+)", handleGetByIdInt)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/item/fortytwo", nil)
	router.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "id is fortytwo")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/item/42", nil)
	router.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusOK)
	data, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert(t, err, nil)
	assert(t, string(data), "id is 42")
}
