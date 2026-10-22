package cmd

import (
	"encoding/json"
	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"reflect"
	"strings"
	"time"
)

var cliRequestSchema, cliResponseSchema bool

type schemaObject = map[string]interface{}

func objectSchema(properties schemaObject, required ...string) schemaObject {
	result := schemaObject{"type": "object", "properties": properties}
	if len(required) > 0 {
		result["required"] = required
	}
	return result
}
func arraySchema(items interface{}) schemaObject {
	return schemaObject{"type": "array", "items": items}
}
func stringSchema() schemaObject { return schemaObject{"type": "string"} }
func dynamicSchema(description string) schemaObject {
	return schemaObject{"type": []string{"object", "array", "string", "number", "boolean", "null"}, "description": description}
}

// jsonTypeSchema derives serialized fields from the same models as the adapters.
// Dynamic upstream extension fields remain allowed rather than being invented.
func jsonTypeSchema(t reflect.Type, seen map[reflect.Type]bool) schemaObject {
	if t.Kind() == reflect.Pointer {
		return jsonTypeSchema(t.Elem(), seen)
	}
	if t == reflect.TypeOf(time.Time{}) {
		return schemaObject{"type": "string", "format": "date-time"}
	}
	if t == reflect.TypeOf(json.RawMessage{}) {
		return dynamicSchema("Raw upstream JSON, defined by the selected operation")
	}
	switch t.Kind() {
	case reflect.String:
		return stringSchema()
	case reflect.Bool:
		return schemaObject{"type": "boolean"}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Uint, reflect.Uint64:
		return schemaObject{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return schemaObject{"type": "number"}
	case reflect.Slice, reflect.Array:
		return schemaObject{"type": []string{"array", "null"}, "items": jsonTypeSchema(t.Elem(), seen)}
	case reflect.Map:
		return schemaObject{"type": []string{"object", "null"}, "additionalProperties": jsonTypeSchema(t.Elem(), seen)}
	case reflect.Struct:
		if seen[t] {
			return schemaObject{"type": "object", "description": "Recursive " + t.Name() + " with the same fields as its parent"}
		}
		seen[t] = true
		defer delete(seen, t)
		props := schemaObject{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if tag == "-" || f.PkgPath != "" {
				continue
			}
			if tag == "" {
				tag = f.Name
			}
			props[tag] = jsonTypeSchema(f.Type, seen)
			if !strings.Contains(f.Tag.Get("json"), "omitempty") {
				omitCustom := t == reflect.TypeOf(models.Document{}) && (tag == "createdAt" || tag == "lastModifiedAt" || tag == "spaceId" || tag == "hasChildren")
				if !omitCustom {
					required = append(required, tag)
				}
			}
		}
		if t == reflect.TypeOf(models.Block{}) {
			props["title"] = schemaObject{"anyOf": []interface{}{stringSchema(), objectSchema(schemaObject{"value": stringSchema()})}}
		}
		return objectSchema(props, required...)
	default:
		return dynamicSchema("Unstructured upstream value")
	}
}
func modelSchema(value interface{}) schemaObject {
	return jsonTypeSchema(reflect.TypeOf(value), map[reflect.Type]bool{})
}

func mcpToolResultSchema() schemaObject {
	return objectSchema(schemaObject{"content": arraySchema(objectSchema(schemaObject{"type": stringSchema(), "text": stringSchema(), "data": stringSchema(), "mimeType": stringSchema()}, "type")), "structuredContent": schemaObject{"type": "object"}, "isError": schemaObject{"type": "boolean"}, "_meta": schemaObject{"type": "object"}}, "content")
}

func cliOutputSchema(c *cobra.Command) schemaObject {
	path := canonicalCommandPath(c)
	var base schemaObject
	switch path {
	case "profiles capabilities":
		base = objectSchema(schemaObject{"name": stringSchema(), "type": stringSchema(), "backends": arraySchema(stringSchema()), "capabilities": arraySchema(stringSchema()), "access_mode": stringSchema(), "permission": stringSchema(), "document_scope": stringSchema()}, "name", "type", "backends", "capabilities")
	case "profiles test":
		base = objectSchema(schemaObject{"name": stringSchema(), "type": stringSchema(), "ok": schemaObject{"type": "boolean"}, "error": stringSchema()}, "name", "type", "ok")
	case "mcp initialize":
		base = objectSchema(schemaObject{"protocolVersion": stringSchema(), "capabilities": schemaObject{"type": "object"}, "serverInfo": objectSchema(schemaObject{"name": stringSchema(), "version": stringSchema()}, "name", "version")}, "protocolVersion", "capabilities", "serverInfo")
	case "mcp tools":
		base = objectSchema(schemaObject{"tools": arraySchema(objectSchema(schemaObject{"name": stringSchema(), "description": stringSchema(), "inputSchema": schemaObject{"type": "object"}}, "name", "inputSchema"))}, "tools")
	case "mcp resources":
		base = objectSchema(schemaObject{"resources": arraySchema(objectSchema(schemaObject{"uri": stringSchema(), "name": stringSchema(), "mimeType": stringSchema()}, "uri", "name"))}, "resources")
	case "mcp edit-review":
		base = objectSchema(schemaObject{"uri": stringSchema(), "name": stringSchema(), "mimeType": stringSchema()}, "uri", "name")
	case "mcp read-resource":
		base = schemaObject{"anyOf": []interface{}{objectSchema(schemaObject{"contents": arraySchema(objectSchema(schemaObject{"uri": stringSchema(), "mimeType": stringSchema(), "text": stringSchema(), "blob": stringSchema()}, "uri"))}, "contents"), objectSchema(schemaObject{"uri": stringSchema(), "name": stringSchema()}, "uri", "name")}}
	case "batch", "mcp batch":
		base = objectSchema(schemaObject{"items": arraySchema(objectSchema(schemaObject{"index": schemaObject{"type": "integer"}, "tool": stringSchema(), "command": stringSchema(), "ok": schemaObject{"type": "boolean"}, "result": dynamicSchema("Negotiated tool result"), "error": stringSchema()}, "index", "tool", "command", "ok")), "total": schemaObject{"type": "integer"}}, "items", "total")
	case "clear":
		base = schemaObject{"anyOf": []interface{}{objectSchema(schemaObject{"items": arraySchema(dynamicSchema("Deleted block outcomes"))}, "items"), objectSchema(schemaObject{"id": stringSchema(), "deleted_blocks": schemaObject{"type": "integer"}}, "id", "deleted_blocks")}}
	case "update":
		base = schemaObject{"anyOf": []interface{}{modelSchema(models.Document{}), objectSchema(schemaObject{"items": arraySchema(dynamicSchema("Block mutation outcomes"))}, "items"), mcpToolResultSchema()}}
	case "limits":
		base = modelSchema(limitsInfo{})
	case "audit agent-dx":
		base = modelSchema(AgentDXAuditResult{})
	case "llm":
		base = modelSchema(llmSpec{})
	case "info":
		base = objectSchema(schemaObject{"connection": modelSchema(models.ConnectionInfo{}), "permissions": schemaObject{"type": "object", "additionalProperties": stringSchema()}}, "connection", "permissions")
	case "cache status":
		base = objectSchema(schemaObject{"documents": schemaObject{"type": "integer"}, "complete": schemaObject{"type": "boolean"}, "updated_at": schemaObject{"type": "string", "format": "date-time"}, "path": stringSchema()}, "documents", "complete", "updated_at", "path")
	case "config migrate":
		base = objectSchema(schemaObject{"status": schemaObject{"const": "migrated"}, "profiles": schemaObject{"type": "integer"}}, "status", "profiles")
	case "cache clear":
		base = objectSchema(schemaObject{"status": schemaObject{"const": "cleared"}}, "status")
	case "get":
		base = modelSchema(models.Document{})
	case "list":
		base = modelSchema(models.DocumentList{})
	case "search":
		base = schemaObject{"anyOf": []interface{}{modelSchema(models.SearchResult{}), modelSchema(models.BlockSearchResultList{})}}
	case "folders list":
		base = modelSchema(models.FolderList{})
	case "tasks list":
		base = modelSchema(models.TaskList{})
	case "collections list":
		base = modelSchema(models.CollectionList{})
	case "collections items":
		base = modelSchema(models.CollectionItemList{})
	case "blocks get":
		base = modelSchema(models.BlocksResponse{})
	case "connection":
		base = modelSchema(models.ConnectionInfo{})
	case "profiles list", "config list":
		base = objectSchema(schemaObject{"items": arraySchema(modelSchema(config.ProfileInfo{})), "source": schemaObject{"const": "config_file"}}, "items", "source")
	case "profiles show":
		base = objectSchema(schemaObject{"profile": modelSchema(config.Profile{}), "source": stringSchema()}, "profile", "source")
	case "version":
		base = objectSchema(schemaObject{"version": stringSchema()}, "version")
	case "sync":
		base = objectSchema(schemaObject{"documents": schemaObject{"type": "integer"}, "updated": schemaObject{"type": "integer"}, "updated_at": schemaObject{"type": "string", "format": "date-time"}, "status": schemaObject{"const": "synced"}}, "documents", "updated", "updated_at", "status")
	case "context", "config inspect":
		base = objectSchema(schemaObject{"version": stringSchema(), "settings": schemaObject{"type": "object", "additionalProperties": objectSchema(schemaObject{"value": dynamicSchema("Effective non-secret setting or credential presence boolean"), "source": stringSchema()}, "value", "source")}, "profiles": arraySchema(modelSchema(config.ProfileInfo{})), "config_path": stringSchema(), "credentials_path": stringSchema(), "scope": schemaObject{"const": "local"}, "rest_available": schemaObject{"type": "boolean"}}, "version", "settings", "profiles", "scope")
	case "local append", "local new", "local open", "local search", "local space", "local today", "local tomorrow", "local yesterday", "setup", "config add", "config add-rest", "config add-mcp", "config use", "config remove", "config reset", "profiles add-rest", "profiles add-mcp", "profiles use", "profiles remove":
		base = objectSchema(schemaObject{"message": stringSchema(), "status": stringSchema()}, "status")
	case "skill-path":
		base = objectSchema(schemaObject{"name": stringSchema(), "source": stringSchema(), "content": stringSchema()}, "name", "source", "content")
	case "feedback":
		base = objectSchema(schemaObject{"path": stringSchema(), "status": stringSchema()}, "path", "status")
	case "validate":
		base = objectSchema(schemaObject{"ok": schemaObject{"type": "boolean"}, "checks": arraySchema(objectSchema(schemaObject{"name": stringSchema(), "ok": schemaObject{"type": "boolean"}}, "name", "ok"))}, "ok", "checks")
	case "completion", "docs", "llm styles", "upgrade":
		base = schemaObject{"type": "string", "contentMediaType": "text/plain", "description": "Human or shell text output; this command has a text stream contract"}
	default:
		if schema, err := restCommandSchema(c, true); err == nil {
			base = schema
			delete(base, "$schema")
		} else {
			base = mcpToolResultSchema()
		}
	}
	if e, ok := declaredEffect(c); ok && e.Effects != "read_only" && e.Scope == "remote" {
		base = schemaObject{"anyOf": []interface{}{base, objectSchema(schemaObject{"operations": arraySchema(base)}, "operations")}}
	}
	if p, _ := restOperation(path); p != "" {
		if spec, err := upstreamRESTSpec(); err == nil {
			base["components"] = spec["components"]
		}
	}
	base["x-craft-command"] = path
	base["x-craft-output-variants"] = schemaObject{"dry_run": objectSchema(schemaObject{"dry_run": schemaObject{"const": true}, "validated": schemaObject{"const": "local"}, "action": stringSchema(), "target": schemaObject{"type": "object"}}, "dry_run", "validated", "action", "target"), "count": objectSchema(schemaObject{"count": schemaObject{"type": "integer"}}, "count"), "projection": "--fields, --transform and --output-only change the selected shape; --format raw preserves upstream bytes; compact may emit an array"}
	base["x-craft-provenance"] = objectSchema(schemaObject{"scope": schemaObject{"enum": []string{"local", "remote"}}, "source": schemaObject{"enum": []string{"local", "live"}}, "backend": stringSchema(), "snapshot_at": stringSchema(), "age_seconds": schemaObject{"type": "integer"}}, "scope", "source", "backend")
	return base
}

func cliInputSchema(c *cobra.Command) schemaObject {
	props := schemaObject{}
	required := []string{}
	add := func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		s := stringSchema()
		switch f.Value.Type() {
		case "bool":
			s = schemaObject{"type": "boolean"}
		case "int", "int64":
			s = schemaObject{"type": "integer"}
		case "stringArray", "stringSlice":
			s = arraySchema(stringSchema())
		}
		s["description"] = f.Usage
		s["x-default"] = f.DefValue
		props[f.Name] = s
		if len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0 {
			required = append(required, f.Name)
		}
	}
	c.InheritedFlags().VisitAll(add)
	c.LocalFlags().VisitAll(add)
	flags := objectSchema(props, required...)
	minArgs, maxArgs := 0, 0
	foundArgs := false
	if c.Args == nil {
		maxArgs = -1
	} else {
		for n := 0; n < 128; n++ {
			args := make([]string, n)
			for i := range args {
				args[i] = "fixture-123"
				if len(c.ValidArgs) > 0 {
					args[i] = c.ValidArgs[0]
				}
			}
			if c.Args(c, args) == nil {
				if !foundArgs {
					minArgs = n
					foundArgs = true
				}
				maxArgs = n
			}
		}
		if maxArgs == 127 {
			maxArgs = -1
		}
	}
	positional := arraySchema(stringSchema())
	positional["minItems"] = minArgs
	if maxArgs >= 0 {
		positional["maxItems"] = maxArgs
	}
	result := objectSchema(schemaObject{"args": positional, "flags": flags})
	result["x-craft-validation"] = "Cobra positional validation and required/mutually-exclusive flag annotations are authoritative; payload validation and upstream capability checks run before mutations"

	conditions := []string{}
	switch canonicalCommandPath(c) {
	case "setup":
		conditions = append(conditions, "Noninteractive: name and exactly one of api-url or mcp-url. Otherwise requires a terminal.")
	case "update":
		conditions = append(conditions, "mode=replace requires yes unless dry-run.")
	case "search":
		conditions = append(conditions, "Provide a query or regex. Exactly one document selects block search.")
	case "upload":
		conditions = append(conditions, "Provide exactly one page/date/sibling placement with compatible position.")
	}
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		for _, key := range []string{"cobra_annotation_mutually_exclusive", "cobra_annotation_required_if_others_set", "cobra_annotation_one_required"} {
			for _, group := range f.Annotations[key] {
				conditions = append(conditions, key+": "+group)
			}
		}
	})
	if len(conditions) > 0 {
		result["x-craft-conditions"] = conditions
	}
	addInputModeContracts(c, result)
	return result
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&cliRequestSchema, "cli-request-schema", false, "Print the CLI argument/flag input contract offline")
	rootCmd.PersistentFlags().BoolVar(&cliResponseSchema, "cli-response-schema", false, "Print the CLI stdout contract offline (distinct from upstream REST)")
}
