# fido

`fido` is a collection of utility functions for constructing HTTP based JSON
APIs with Go. It is dependency free.

`fido` includes:
- A simple method based regexp router.
- A suite of helper functions for writing and binding JSON.
- A small suite of middleware.

`fido` is designed to work closely with `net/http` stdlib functions and types.

## Installation

```
go get github.com/dbridges/fido
```

## Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/dbridges/fido"
)

func handleSay(w http.ResponseWriter, r *http.Request) {
	// Params returns an object you can query for named path parameters
	msg := fido.Params(r).Get("msg")
	fido.JSON(w, http.StatusOK, fido.H{"message": msg + ", " + msg})
}

func main() {
	r := fido.NewRouter()

	// Add some basic middleware
	r.Use(fido.Logger)
	r.Use(fido.Recoverer)

	// Handle takes an http.Handler or a function to be cast to
	// http.HandlerFunc.
	//
	// Paths can include regular Go regexps. Named capture groups will be
	// available in the request's context.
	r.Handle("GET", "/say/(?P<msg>[a-zA-Z]+)", handleSay)

	log.Println("Listening on port 5000")
	log.Fatal(http.ListenAndServe("localhost:5000", r))
}
```
