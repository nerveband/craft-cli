// Package contracts contains the captured public API contract used for offline discovery.
package contracts

import (
	"embed"
	"encoding/json"
)

//go:embed craft-rest-space-openapi.json
var files embed.FS

func OpenAPI() (map[string]interface{}, error) {
	data, err := files.ReadFile("craft-rest-space-openapi.json")
	if err != nil {
		return nil, err
	}
	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	return spec, err
}
