// Package apidoc serves the native Go HTTP contract.
package apidoc

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"sync"
)

//go:embed openapi-template.json
var schemaTemplate []byte

type jsonObject map[string]json.RawMessage

var schema = sync.OnceValues(func() ([]byte, error) {
	var document jsonObject
	if err := json.Unmarshal(schemaTemplate, &document); err != nil {
		return nil, err
	}
	var common jsonObject
	if err := json.Unmarshal(document["commonResponses"], &common); err != nil {
		return nil, err
	}
	delete(document, "commonResponses")

	var paths map[string]map[string]jsonObject
	if err := json.Unmarshal(document["paths"], &paths); err != nil {
		return nil, err
	}
	for _, methods := range paths {
		for _, operation := range methods {
			responses := make(jsonObject, len(common)+1)
			for status, response := range common {
				responses[status] = response
			}
			// Operation-specific responses override shared defaults.
			if err := json.Unmarshal(operation["responses"], &responses); err != nil {
				return nil, err
			}
			data, err := json.Marshal(responses)
			if err != nil {
				return nil, err
			}
			operation["responses"] = data
		}
	}
	data, err := json.Marshal(paths)
	if err != nil {
		return nil, err
	}
	document["paths"] = data
	return json.Marshal(document)
})

func ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	data, err := schema()
	if err != nil {
		http.Error(w, "schema unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}
