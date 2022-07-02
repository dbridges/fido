package fido

import (
	"log"
	"net/http"
	"time"
)

// Middleware defines a standard http middleware function which takes in an
// http.Handler and returns a new http.Handler.
type Middleware func(http.Handler) http.Handler

// Recoverer is a middleware which provides basic panic recovery. If a panic
// occurs Recoverer will recover from it and write an internal server error
// message to the response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				JSONError(w, http.StatusInternalServerError, "an unknown error occured")
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Logger is a middleware which provides basic request logging.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		log.Printf("%-5s %s in %s\n", r.Method, r.URL, end.Sub(start))
	})
}

// BasicAuth is a middleware the provides HTTP basic auth protection. A verify
// function is requireed which should return true if the supplied username and
// password are authorized. If credentials are not provided a status bad
// request error is written to the response. If the verify function returns
// false a status unauthorized error is written to the response.
func BasicAuth(verify func(u, p string) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				JSONError(w, http.StatusBadRequest, "Authorization requiried")
				return
			}
			if !verify(u, p) {
				JSONError(w, http.StatusUnauthorized, "Invalid username or password")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
