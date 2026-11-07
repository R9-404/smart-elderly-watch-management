package apidoc

import (
	"encoding/json"
	"testing"
)

func TestRiskDeleteContractHasPathAndCSRFWithoutBody(t *testing.T) {
	b, err := buildSchema()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			RequestBody json.RawMessage `json:"requestBody"`
			Parameters  []struct {
				Name     string `json:"name"`
				In       string `json:"in"`
				Required bool   `json:"required"`
				Ref      string `json:"$ref"`
			} `json:"parameters"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	op, ok := doc.Paths["/api/v1/risk/rules/{id}"]["delete"]
	if !ok || len(op.RequestBody) != 0 {
		t.Fatal("delete contract requires nonexistent JSON body")
	}
	path, csrf := false, false
	for _, p := range op.Parameters {
		path = path || (p.Name == "id" && p.In == "path" && p.Required)
		csrf = csrf || p.Ref == "#/components/parameters/CsrfToken"
	}
	if !path || !csrf {
		t.Fatal("missing path or CSRF parameter")
	}
}
