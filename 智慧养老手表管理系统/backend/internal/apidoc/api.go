// Package apidoc serves the native Go HTTP contract.
package apidoc

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var schema []byte

func ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(schema)
}
