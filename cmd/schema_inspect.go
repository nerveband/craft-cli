package cmd

import (
	"fmt"
	"strings"

	"github.com/ashrafali/craft-cli/docs/contracts"
	"github.com/spf13/cobra"
)

// Inspect schema requests before Cobra validates resource arguments or initializes config.
func inspectSchemaArgs(args []string) (bool, error) {
	requested := false
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--request-schema" || arg == "--response-schema" || arg == "--request-schema=true" || arg == "--response-schema=true" {
			requested = true
		}
	}
	if !requested {
		return false, nil
	}
	command, flags, err := rootCmd.Find(args)
	if err != nil {
		return true, err
	}
	if err := command.ParseFlags(flags); err != nil {
		return true, err
	}
	if requestSchema && responseSchema {
		return true, fmt.Errorf("choose --request-schema or --response-schema")
	}
	if !requestSchema && !responseSchema {
		return false, nil
	}
	schema, err := restCommandSchema(command, responseSchema)
	if err != nil {
		return true, err
	}
	return true, outputSchemaJSON(schema)
}

func restOperation(path string) (string, string) {
	operations := map[string][2]string{
		"list": {"/documents", "get"}, "get": {"/blocks", "get"}, "create": {"/documents", "post"}, "delete": {"/documents", "delete"}, "move": {"/documents/move", "put"},
		"blocks get": {"/blocks", "get"}, "blocks add": {"/blocks", "post"}, "blocks update": {"/blocks", "put"}, "blocks delete": {"/blocks", "delete"}, "blocks move": {"/blocks/move", "put"},
		"search": {"/documents/search", "get"}, "tasks list": {"/tasks", "get"}, "tasks add": {"/tasks", "post"}, "tasks update": {"/tasks", "put"}, "tasks delete": {"/tasks", "delete"},
		"folders list": {"/folders", "get"}, "folders create": {"/folders", "post"}, "folders delete": {"/folders", "delete"}, "folders move": {"/folders/move", "put"},
		"collections list": {"/collections", "get"}, "collections create": {"/collections", "post"}, "collections schema": {"/collections/{collectionId}/schema", "get"}, "collections schema update": {"/collections/{collectionId}/schema", "put"},
		"collections items": {"/collections/{collectionId}/items", "get"}, "collections add": {"/collections/{collectionId}/items", "post"}, "collections update": {"/collections/{collectionId}/items", "put"}, "collections delete": {"/collections/{collectionId}/items", "delete"},
		"collections views list": {"/collections/{collectionId}/views", "get"}, "collections views create": {"/collections/{collectionId}/views", "post"}, "collections views update": {"/collections/{collectionId}/views/{viewId}", "put"}, "collections views delete": {"/collections/{collectionId}/views/{viewId}", "delete"}, "collections active-view set": {"/collections/{collectionId}/active-view", "put"},
		"reminders list": {"/reminders", "get"}, "reminders create": {"/reminders", "post"}, "reminders update": {"/reminders", "put"}, "reminders delete": {"/reminders", "delete"},
		"upload": {"/upload", "post"}, "comments add": {"/comments", "post"}, "connection": {"/connection", "get"}, "whiteboards create": {"/whiteboards", "post"}, "whiteboards get": {"/whiteboards/{whiteboardBlockId}/elements", "get"}, "whiteboards add": {"/whiteboards/{whiteboardBlockId}/elements", "post"}, "whiteboards update": {"/whiteboards/{whiteboardBlockId}/elements", "put"}, "whiteboards delete": {"/whiteboards/{whiteboardBlockId}/elements", "delete"},
	}
	op := operations[strings.TrimPrefix(path, "craft ")]
	return op[0], op[1]
}

func restCommandSchema(command *cobra.Command, response bool) (map[string]interface{}, error) {
	path, method := restOperation(command.CommandPath())
	if path == "" {
		return nil, fmt.Errorf("REST schema is unavailable for %s; use craft schema --command %q for CLI discovery", command.CommandPath(), strings.TrimPrefix(command.CommandPath(), "craft "))
	}
	spec, err := contracts.OpenAPI()
	if err != nil {
		return nil, err
	}
	operation := spec["paths"].(map[string]interface{})[path].(map[string]interface{})[method].(map[string]interface{})
	var schema map[string]interface{}
	if response {
		schema = operation["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
	} else if body, ok := operation["requestBody"].(map[string]interface{}); ok {
		contents := body["content"].(map[string]interface{})
		if jsonBody, ok := contents["application/json"].(map[string]interface{}); ok {
			schema = jsonBody["schema"].(map[string]interface{})
		} else {
			schema = map[string]interface{}{"type": "string", "format": "binary"}
		}
	} else {
		props := map[string]interface{}{}
		required := []string{}
		params, _ := operation["parameters"].([]interface{})
		for _, item := range params {
			p := item.(map[string]interface{})
			name := p["name"].(string)
			props[name] = p["schema"]
			if p["required"] == true {
				required = append(required, name)
			}
		}
		schema = map[string]interface{}{"type": "object", "properties": props, "required": required}
	}
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["x-craft-cli-contract"] = "upstream REST payload; CLI convenience output may add or flatten fields"
	schema["x-craft-operation"] = strings.ToUpper(method) + " " + path
	if components, ok := spec["components"]; ok {
		schema["components"] = components
	}
	return schema, nil
}
