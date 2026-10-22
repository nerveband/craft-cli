package cmd

import "github.com/spf13/cobra"

// These rules describe input modes that Cobra's positional callback selects at
// runtime. The same callback remains the execution gate; schema probes test both modes.
func addInputModeContracts(c *cobra.Command, schema schemaObject) {
	properties := schema["properties"].(schemaObject)
	flags := properties["flags"].(schemaObject)
	flagProperties := flags["properties"].(schemaObject)
	flags["additionalProperties"] = false
	present := func(name string) schemaObject { return objectSchema(schemaObject{name: flagProperties[name]}, name) }
	enabled := func(name string) schemaObject {
		return objectSchema(schemaObject{name: schemaObject{"const": true}}, name)
	}
	sources := []interface{}{}
	for _, name := range []string{"json", "json-file", "stdin"} {
		if _, ok := flagProperties[name]; ok {
			if name == "stdin" {
				sources = append(sources, enabled(name))
			} else {
				sources = append(sources, present(name))
			}
		}
	}
	path := canonicalCommandPath(c)
	if len(sources) > 0 {
		// Structured input is exclusive, but commands such as blocks add may also
		// take a target ID while supplying JSON. Only raw whole-request modes replace args.
		if len(sources) > 1 {
			flags["allOf"] = []interface{}{schemaObject{"not": schemaObject{"anyOf": inputSourcePairs(sources)}}}
		}
		rawReplacesArgs := false
		switch path {
		case "create", "delete", "move", "folders create", "folders move", "folders delete", "tasks add", "tasks update", "tasks delete", "whiteboards create", "whiteboards delete":
			rawReplacesArgs = true
		}
		if rawReplacesArgs {
			positional := properties["args"].(schemaObject)
			normalMin := 1
			if path == "create" {
				normalMin = 0
			}
			positional["minItems"] = 0
			schema["allOf"] = []interface{}{schemaObject{"if": objectSchema(schemaObject{"flags": schemaObject{"anyOf": sources}}, "flags"), "then": objectSchema(schemaObject{"args": schemaObject{"maxItems": 0}}), "else": objectSchema(schemaObject{"args": schemaObject{"minItems": normalMin}})}}
		}
	}
	requiredWhenNotRaw := map[string][]string{
		"folders move": {"to"}, "collections rename": {"name"}, "collections add": {"title"}, "collections update": {"item"}, "collections delete": {"item"}, "collections views create": {"name", "type"}, "collections views update": {"view"}, "collections views delete": {"view"}, "collections active-view set": {"view"}, "blocks move": {"to"},
	}
	if required, ok := requiredWhenNotRaw[path]; ok {
		condition := schemaObject{"required": required}
		if len(sources) > 0 {
			condition = schemaObject{"anyOf": []interface{}{schemaObject{"anyOf": sources}, condition}}
		}
		previous, _ := flags["allOf"].([]interface{})
		flags["allOf"] = append(previous, condition)
	}
	if path == "setup" {
		rest := objectSchema(schemaObject{"name": stringSchema(), "api-url": stringSchema()}, "name", "api-url")
		mcp := objectSchema(schemaObject{"name": stringSchema(), "mcp-url": stringSchema()}, "name", "mcp-url")
		flags["oneOf"] = []interface{}{schemaObject{"not": schemaObject{"anyOf": []interface{}{schemaObject{"required": []string{"name"}}, schemaObject{"required": []string{"api-url"}}, schemaObject{"required": []string{"mcp-url"}}}}}, rest, mcp}
	}
}
func inputSourcePairs(sources []interface{}) []interface{} {
	pairs := []interface{}{}
	for i := 0; i < len(sources); i++ {
		for j := i + 1; j < len(sources); j++ {
			pairs = append(pairs, schemaObject{"allOf": []interface{}{sources[i], sources[j]}})
		}
	}
	return pairs
}
