package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func runCollectionView(action, collectionID string) error {
	if err := validateResourceID(collectionID, "collection-id"); err != nil {
		return err
	}
	if action == "update" || action == "delete" || action == "set-active" {
		if err := validateResourceID(collectionViewID, "view-id"); err != nil {
			return err
		}
	}
	path := "/collections/" + url.PathEscape(collectionID) + "/views"
	method := "GET"
	var payload map[string]interface{}
	if action == "create" || action == "update" {
		if collectionJSON != "" || collectionStdin {
			var err error
			payload, err = readRawPayload(collectionJSON, collectionStdin)
			if err != nil {
				return err
			}
		} else {
			view := map[string]interface{}{}
			if collectionViewName != "" {
				view["name"] = collectionViewName
			}
			if collectionViewType != "" {
				view["type"] = collectionViewType
			}
			payload = map[string]interface{}{"view": view}
		}
		view, ok := payload["view"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("payload requires a view object")
		}
		if action == "create" && (view["name"] == nil || view["type"] == nil) {
			return fmt.Errorf("create view requires --name and --type, or a JSON view with name and type")
		}
		if t, ok := view["type"]; ok && t != "table" && t != "gallery" && t != "kanban" {
			return fmt.Errorf("invalid view type %v; valid: table, gallery, kanban", t)
		}
		if groups, ok := view["groupBy"].([]interface{}); ok && len(groups) > 1 {
			return fmt.Errorf("view groupBy supports at most one rule")
		}
		if view["type"] == "kanban" {
			groups, _ := view["groupBy"].([]interface{})
			if len(groups) != 1 {
				return fmt.Errorf("kanban requires exactly one groupBy rule in --json")
			}
		}
		for field, limit := range map[string]int{"filters": 5, "sortBy": 3} {
			if list, ok := view[field].([]interface{}); ok && len(list) > limit {
				return fmt.Errorf("%s supports at most %d rules", field, limit)
			}
		}
		method = "POST"
		if action == "update" {
			method = "PUT"
			path += "/" + url.PathEscape(collectionViewID)
		}
	}
	if action == "delete" {
		method = "DELETE"
		path += "/" + url.PathEscape(collectionViewID)
	}
	if action == "set-active" {
		method = "PUT"
		path = "/collections/" + url.PathEscape(collectionID) + "/active-view"
		payload = map[string]interface{}{"viewId": collectionViewID}
	}
	if backendName == "mcp" || selectedMCPProfile() || collectionDiff || collectionSaveRevert != "" {
		command := "collections views-" + action + " --collection " + quoteMCPArg(collectionID)
		if action != "list" && action != "create" {
			command += " --view " + quoteMCPArg(collectionViewID)
		}
		if payload != nil && action != "set-active" {
			view := payload["view"].(map[string]interface{})
			// These mappings are captured from craft_write collections views-create --help.
			flags := map[string]string{"name": "name", "type": "type", "filters": "filters", "sortBy": "sort", "groupBy": "group", "calculations": "calculations", "isCalculationsRowVisible": "show-calculations"}
			for key, value := range view {
				flag, ok := flags[key]
				if !ok {
					return fmt.Errorf("view field %q has no verified MCP mapping; use --backend rest", key)
				}
				text, ok := value.(string)
				if !ok {
					b, _ := json.Marshal(value)
					text = string(b)
				}
				command += " --" + flag + " " + quoteMCPArg(text)
			}
		}
		tool := "craft_write"
		if action == "list" {
			tool = "craft_read"
		}
		return runMCPCollectionCommand("collections.views."+action, command, tool)
	}
	if isDryRun() {
		return dryRunOutput("collections views "+action, map[string]interface{}{"method": method, "path": path, "payload": payload, "reversible": false})
	}
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	result, err := client.RequestJSON(method, path, payload)
	if err != nil {
		return err
	}
	return outputRawJSON(result)
}
