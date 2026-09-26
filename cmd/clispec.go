package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

func specArgs(flags *pflag.FlagSet) []map[string]interface{} {
	result := []map[string]interface{}{}
	flags.VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		typ := f.Value.Type()
		switch typ {
		case "bool":
			typ = "boolean"
		case "int", "int64":
			typ = "integer"
		case "stringArray", "stringSlice":
			typ = "string[]"
		case "intSlice":
			typ = "integer[]"
		}
		a := map[string]interface{}{"name": "--" + f.Name, "type": typ, "description": f.Usage, "required": len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0}
		if f.Shorthand != "" {
			a["short"] = "-" + f.Shorthand
		}
		result = append(result, a)
	})
	return result
}

// cliSpec publishes the frozen 0.2 schema with additional explicit effects.
// Behavioral compliance is evaluated separately by Audit v3.
func cliSpec(filter string) map[string]interface{} {
	commands := []map[string]interface{}{}
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		for _, c := range parent.Commands() {
			if c.Name() == "help" {
				continue
			}
			path := strings.TrimPrefix(c.CommandPath(), "craft ")
			e, ok := declaredEffect(c)
			if ok && c.Runnable() && (filter == "" || path == filter || strings.HasPrefix(path, filter+" ")) {
				args := specArgs(c.LocalNonPersistentFlags())
				for _, token := range strings.Fields(c.Use)[1:] {
					if strings.HasPrefix(token, "[") || strings.HasPrefix(token, "<") {
						args = append(args, map[string]interface{}{"name": strings.Trim(token, "[]<>"), "type": "string", "required": strings.HasPrefix(token, "<")})
					}
				}
				entry := map[string]interface{}{"name": path, "description": c.Short, "effects": e.Effects, "mutating": e.Effects != "read_only", "stdout_schema": map[string]interface{}{}, "args": args, "extensions": map[string]interface{}{"scope": e.Scope, "destructive": e.Destructive, "supports_dry_run": e.Effects != "read_only", "retry_guidance": "After a write timeout verify with get/list/search before retrying; the server may have applied it.", "output_contract": "Use --response-schema for the upstream REST shape; convenience commands may flatten it."}}
				if e.Destructive {
					entry["confirmation_bypass_arg"] = "--yes"
				}
				if path == "update" {
					entry["confirmation_bypass_arg"] = "--yes"
					entry["extensions"].(map[string]interface{})["confirmation_condition"] = "Only --mode replace requires --yes; the default append mode does not."
				}
				if path == "setup" {
					entry["requires_tty"] = true
				}
				// Do not claim bounded cardinality for collection APIs without server paging.
				if path == "list" || path == "folders list" || path == "tasks list" || path == "collections list" {
					entry["extensions"].(map[string]interface{})["pagination_limitation"] = "REST returns the full collection; limits and counts are computed locally."
				}
				if c.Example != "" {
					entry["extensions"].(map[string]interface{})["examples"] = strings.Split(strings.TrimSpace(c.Example), "\n")
				}
				commands = append(commands, entry)
			}
			walk(c)
		}
	}
	walk(rootCmd)
	errors := []map[string]interface{}{}
	for _, kind := range []string{"USER_ERROR", "CONFIG_ERROR", "API_ERROR", "API_TIMEOUT", "AUTH_ERROR", "PERMISSION_DENIED", "NOT_FOUND", "RATE_LIMIT", "PAYLOAD_TOO_LARGE", "MCP_TOOL_ERROR", "PARTIAL_FAILURE", "CONFIRMATION_REQUIRED", "CAPABILITY_UNAVAILABLE", "BACKEND_UNAVAILABLE", "READ_UNAVAILABLE", "WRITE_UNAVAILABLE"} {
		code := 1
		switch kind {
		case "CONFIG_ERROR":
			code = 3
		case "API_ERROR", "API_TIMEOUT", "AUTH_ERROR", "PERMISSION_DENIED", "NOT_FOUND", "RATE_LIMIT", "PAYLOAD_TOO_LARGE", "MCP_TOOL_ERROR", "PARTIAL_FAILURE":
			code = 2
		}
		errors = append(errors, map[string]interface{}{"kind": strings.ToLower(kind), "exit_code": code, "retryable": kind == "RATE_LIMIT" || kind == "API_ERROR"})
	}
	return map[string]interface{}{"clispec": "0.2", "command_layout": "flat", "name": "craft", "version": version, "description": rootCmd.Short, "output": map[string]string{"tty": "json", "piped": "json"}, "global_args": specArgs(rootCmd.PersistentFlags()), "commands": commands, "errors": errors, "extensions": map[string]interface{}{"generated_from": "Cobra command definitions and explicit cmd/effects.go declarations; do not hand-edit docs/command-reference.json", "status": "v0.2 schema; see Audit v3 final report for behavioral gaps", "async": "No asynchronous submission or polling API is exposed; no job ledger or wait is applicable.", "feedback": "Local feedback only; no upstream feedback transport.", "profiles": "Discover safe runtime profile metadata with profiles list; schema never reads user configuration."}}
}
