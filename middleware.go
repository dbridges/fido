package fido

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

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

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		log.Printf("%-5s %s in %s\n", r.Method, r.URL, end.Sub(start))
	})
}

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
