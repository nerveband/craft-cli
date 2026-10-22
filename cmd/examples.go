package cmd

import (
	"github.com/spf13/cobra"
	"strings"
)

// These are executable CLI examples with synthetic IDs, never user data.
var workflowExamples = map[string]string{
	"audit agent-dx":              "craft audit agent-dx --checks",
	"batch":                       "craft batch --file operations.json --dry-run",
	"blocks add":                  "craft blocks add page-123 --markdown 'Meeting notes' --dry-run",
	"blocks delete":               "craft blocks delete block-123 --dry-run",
	"blocks explore-themes":       "craft blocks explore-themes --type page",
	"blocks explore-washi":        "craft blocks explore-washi",
	"blocks get":                  "craft blocks get page-123 --depth 1",
	"blocks learn":                "craft blocks learn pages tables",
	"blocks move":                 "craft blocks move block-123 --to page-456 --dry-run",
	"blocks revert":               "craft blocks revert --revert-info-file revert.json --dry-run",
	"blocks search-unsplash":      "craft blocks search-unsplash 'mountain lake'",
	"blocks update":               "craft blocks update block-123 --markdown 'Revised notes' --dry-run",
	"clear":                       "craft clear page-123 --dry-run",
	"collections active-view set": "craft collections active-view set collection-123 --view view-456 --dry-run",
	"collections add":             "craft collections add collection-123 --title 'Release checklist' --properties '{\"Status\":\"Todo\"}' --dry-run",
	"collections create":          "craft collections create --json '{\"schema\":{\"name\":\"Projects\",\"properties\":[]},\"position\":{\"pageId\":\"page-123\",\"position\":\"end\"}}' --dry-run",
	"collections delete":          "craft collections delete collection-123 --item item-456 --dry-run",
	"collections items":           "craft collections items collection-123 --depth 1",
	"collections list":            "craft collections list --document page-123",
	"collections rename":          "craft collections rename collection-123 --name 'Active projects' --dry-run",
	"collections schema":          "craft collections schema collection-123",
	"collections schema update":   "craft collections schema update collection-123 --json '{\"schema\":{\"name\":\"Projects\",\"properties\":[]}}' --dry-run",
	"collections update":          "craft collections update collection-123 --item item-456 --properties '{\"Status\":\"Done\"}' --dry-run",
	"collections views create":    "craft collections views create collection-123 --name Overview --type table --dry-run",
	"collections views delete":    "craft collections views delete collection-123 --view view-456 --dry-run",
	"collections views list":      "craft collections views list collection-123",
	"collections views update":    "craft collections views update collection-123 --view view-456 --name 'Team board' --dry-run",
	"comments add":                "craft comments add --json '{\"comments\":[{\"blockId\":\"block-123\",\"content\":\"Reviewed\"}]}' --dry-run",
	"completion":                  "craft completion zsh",
	"config add":                  "craft config add work https://connect.craft.do/links/demo/api/v1 --dry-run",
	"config add-mcp":              "craft config add-mcp work-mcp --mcp-url https://mcp.craft.do/links/demo/mcp --dry-run",
	"config add-rest":             "craft config add-rest work --api-url https://connect.craft.do/links/demo/api/v1 --api-key-env CRAFT_API_KEY --dry-run",
	"config list":                 "craft config list",
	"config inspect":              "craft config inspect --profile work",
	"config remove":               "craft config remove retired --dry-run",
	"config reset":                "craft config reset --dry-run",
	"config use":                  "craft config use work --dry-run",
	"connection":                  "craft connection",
	"context":                     "craft context --profile work",
	"create":                      "craft create --title 'Weekly review' --markdown 'Agenda' --dry-run",
	"delete":                      "craft delete page-123 --dry-run",
	"docs":                        "craft docs",
	"documents resolve-link":      "craft documents resolve-link 'craftdocs://open?spaceId=space-123&documentId=document-456'",
	"feedback":                    "craft feedback 'Search returned an unexpected field' --dry-run",
	"folders create":              "craft folders create Projects --dry-run",
	"folders delete":              "craft folders delete folder-123 --dry-run",
	"folders explore-icons":       "craft folders explore-icons folder",
	"folders list":                "craft folders list",
	"folders move":                "craft folders move folder-123 --to folder-456 --dry-run",
	"get":                         "craft get page-123 --max-depth 1",
	"images view":                 "craft images view https://example.org/photo.png",
	"info":                        "craft info",
	"limits":                      "craft limits",
	"list":                        "craft list --limit 10 --metadata",
	"llm":                         "craft llm",
	"llm styles":                  "craft llm styles",
	"local append":                "craft local append page-123 'Review notes' --space-id space-123 --dry-run",
	"local new":                   "craft local new --title 'Weekly review' --space-id space-123 --dry-run",
	"local open":                  "craft local open page-123 --space-id space-123 --dry-run",
	"local search":                "craft local search roadmap --space-id space-123 --dry-run",
	"local space":                 "craft local space --space-id space-123 --dry-run",
	"local today":                 "craft local today --space-id space-123 --dry-run",
	"local tomorrow":              "craft local tomorrow --space-id space-123 --dry-run",
	"local yesterday":             "craft local yesterday --space-id space-123 --dry-run",
	"mcp batch":                   "craft mcp batch --file commands.json --dry-run",
	"mcp call":                    "craft mcp call craft_read --command 'documents list --limit 5'",
	"mcp edit-review":             "craft mcp edit-review",
	"mcp initialize":              "craft mcp initialize",
	"mcp read-resource":           "craft mcp read-resource ui://craft/edit-review --metadata-only",
	"mcp resources":               "craft mcp resources",
	"mcp tools":                   "craft mcp tools",
	"move":                        "craft move page-123 --to-location unsorted --dry-run",
	"profiles add-mcp":            "craft profiles add-mcp work-mcp --mcp-url https://mcp.craft.do/links/demo/mcp --dry-run",
	"profiles add-rest":           "craft profiles add-rest work --api-url https://connect.craft.do/links/demo/api/v1 --api-key-env CRAFT_API_KEY --dry-run",
	"profiles capabilities":       "craft profiles capabilities work",
	"profiles list":               "craft profiles list",
	"profiles remove":             "craft profiles remove retired --dry-run",
	"profiles show":               "craft profiles show work",
	"profiles test":               "craft profiles test work",
	"profiles use":                "craft profiles use work --dry-run",
	"reminders create":            "craft reminders create block-123 --dry-run",
	"reminders delete":            "craft reminders delete reminder-123 --dry-run",
	"reminders list":              "craft reminders list --status incomplete --limit 10",
	"reminders update":            "craft reminders update reminder-123 --completed --dry-run",
	"schema":                      "craft schema --command 'blocks update'",
	"search":                      "craft search 'release notes' --limit 5",
	"setup":                       "craft setup --name work --api-url https://connect.craft.do/links/demo/api/v1 --api-key-env CRAFT_API_KEY --dry-run",
	"skill-path":                  "craft skill-path",
	"sync":                        "craft sync --full --dry-run",
	"tasks add":                   "craft tasks add 'Review release' --dry-run",
	"tasks delete":                "craft tasks delete task-123 --dry-run",
	"tasks list":                  "craft tasks list --scope inbox",
	"tasks update":                "craft tasks update task-123 --state done --dry-run",
	"update":                      "craft update page-123 --markdown 'Additional notes' --dry-run",
	"upgrade":                     "craft upgrade --dry-run",
	"upload":                      "craft upload report.pdf --page page-123 --dry-run",
	"validate":                    "craft validate",
	"version":                     "craft version",
	"whiteboards add":             "craft whiteboards add whiteboard-123 --json '[{\"id\":\"shape-123\",\"type\":\"rectangle\",\"x\":0,\"y\":0,\"width\":100,\"height\":100}]' --dry-run",
	"whiteboards create":          "craft whiteboards create page-123 --dry-run",
	"whiteboards delete":          "craft whiteboards delete whiteboard-123 --ids shape-123 --dry-run",
	"whiteboards elements get":    "craft whiteboards elements get whiteboard-123",
	"whiteboards get":             "craft whiteboards get whiteboard-123",
	"whiteboards update":          "craft whiteboards update whiteboard-123 --json '[{\"id\":\"shape-123\",\"x\":100}]' --dry-run",
}

func installExamples() {
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if example, ok := workflowExamples[canonicalCommandPath(c)]; ok {
			if strings.HasPrefix(c.CommandPath(), "craft documents ") && canonicalCommandPath(c) != strings.TrimPrefix(c.CommandPath(), "craft ") {
				example = strings.Replace(example, "craft ", "craft documents ", 1)
			}
			c.Example = "  " + example
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}
