package fido

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRecoverer(t *testing.T) {
	handlePanic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error")
	})
	withRecoverer := Recoverer(handlePanic)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	withRecoverer.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusInternalServerError)
}

func TestLogger(t *testing.T) {
	var output bytes.Buffer
	log.SetOutput(&output)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	handleLogger := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withLogger := Logger(handleLogger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	withLogger.ServeHTTP(w, r)
	assert(t, strings.Index(output.String(), "GET     /path") > 0, true)
}

func TestBasicAuth(t *testing.T) {
	verify := func(u, p string) bool {
		return u == "user" && p == "password"
	}

	handleBasicAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withBasicAuth := BasicAuth(verify)(handleBasicAuth)

	// Test returns bad request without basic auth headers
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	withBasicAuth.ServeHTTP(w, r)
	res := w.Result()
	assert(t, res.StatusCode, http.StatusBadRequest)

	// Test returns unauthorized if provided username doesn't match.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("not right", "password")
	withBasicAuth.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusUnauthorized)

	// Test returns unauthorized if provided password doesn't match.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("user", "not right")
	withBasicAuth.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusUnauthorized)

	// Test returns okay if provided username and password match.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("user", "password")
	withBasicAuth.ServeHTTP(w, r)
	res = w.Result()
	assert(t, res.StatusCode, http.StatusOK)
}
