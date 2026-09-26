package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"strconv"
	"time"
)

var remindersCmd = &cobra.Command{Use: "reminders", Short: "Manage reminders and Save for later"}

func init() {
	rootCmd.AddCommand(remindersCmd)
	commandEffects["reminders"] = CommandEffect{"read_only", "remote", false}
	for _, action := range []string{"list", "create", "update", "delete"} {
		action := action
		var raw, at, status, cursor string
		var stdin, clearTime, completed bool
		var limit int
		c := &cobra.Command{Use: action + " [id]", Short: action + " reminders", Args: cobra.MaximumNArgs(1), Example: "  craft reminders list --limit 10\n  craft reminders create BLOCK_ID --at 2030-01-15T10:00:00-05:00 --dry-run", RunE: func(cmd *cobra.Command, args []string) error {
			method := map[string]string{"list": "GET", "create": "POST", "update": "PUT", "delete": "DELETE"}[action]
			path := "/reminders"
			var payload map[string]interface{}
			if action == "list" {
				if len(args) > 0 {
					return fmt.Errorf("reminders list accepts no ID")
				}
				if limit < 1 || limit > 200 {
					return fmt.Errorf("--limit must be between 1 and 200")
				}
				q := url.Values{}
				if status != "" {
					if status != "incomplete" && status != "upcoming" && status != "completed" && status != "all" {
						return fmt.Errorf("invalid status %q; valid: incomplete, completed, upcoming, all", status)
					}
					q.Set("status", status)
				}
				if limit > 0 {
					q.Set("limit", strconv.Itoa(limit))
				}
				if cursor != "" {
					q.Set("cursor", cursor)
				}
				if len(q) > 0 {
					path += "?" + q.Encode()
				}
			} else if raw != "" || stdin {
				var err error
				payload, err = readRawPayload(raw, stdin)
				if err != nil {
					return err
				}
			} else {
				if len(args) != 1 {
					return fmt.Errorf("%s requires an ID or --json/--stdin payload", action)
				}
				if err := validateResourceID(args[0], "id"); err != nil {
					return err
				}
				if at != "" {
					parsed, err := time.Parse(time.RFC3339, at)
					if err != nil {
						return fmt.Errorf("--at requires RFC3339 with Z or UTC offset")
					}
					if !parsed.After(time.Now()) {
						return fmt.Errorf("--at must be in the future; include the intended date's timezone offset")
					}
				}
				if clearTime && at != "" {
					return fmt.Errorf("choose --at or --clear-time")
				}
				item := map[string]interface{}{}
				if at != "" {
					item["remindAt"] = at
				}
				if clearTime {
					item["remindAt"] = nil
				}
				switch action {
				case "create":
					item["blockId"] = args[0]
					payload = map[string]interface{}{"reminders": []interface{}{item}}
				case "update":
					item["id"] = args[0]
					if cmd.Flags().Changed("completed") {
						item["isCompleted"] = completed
					}
					if len(item) == 1 {
						return fmt.Errorf("update requires --at, --clear-time or --completed")
					}
					payload = map[string]interface{}{"remindersToUpdate": []interface{}{item}}
				case "delete":
					payload = map[string]interface{}{"idsToDelete": args}
				}
			}
			if isDryRun() {
				return dryRunOutput("reminders "+action, map[string]interface{}{"method": method, "path": path, "payload": payload, "reversible": false})
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
		}}
		if action == "list" {
			c.Flags().StringVar(&status, "status", "", "Filter: incomplete, completed, upcoming, all")
			c.Flags().IntVar(&limit, "limit", 50, "Server page size")
			c.Flags().StringVar(&cursor, "cursor", "", "Server pagination cursor")
			commandEffects["reminders list"] = CommandEffect{"read_only", "remote", false}
		} else {
			c.Flags().StringVar(&raw, "json", "", "Exact REST JSON payload")
			c.Flags().BoolVar(&stdin, "stdin", false, "Read exact REST payload from stdin")
			if action != "delete" {
				c.Flags().StringVar(&at, "at", "", "Future RFC3339 timestamp; omit on create for Save for later")
				c.Flags().BoolVar(&clearTime, "clear-time", false, "Clear notification time, keeping Save for later")
			}
			if action == "update" {
				c.Flags().BoolVar(&completed, "completed", false, "Set completion explicitly (use --completed=false to reopen)")
			}
			commandEffects["reminders "+action] = CommandEffect{"non_idempotent", "remote", action == "delete"}
		}
		remindersCmd.AddCommand(c)
	}
}
