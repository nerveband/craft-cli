package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
	Long: `Manage Craft tasks - list, add, update, and delete tasks.

Examples:
  craft tasks list                           # List all tasks
  craft tasks list --scope active            # List active tasks
  craft tasks list --document ID             # List tasks in document
  craft tasks add "Buy groceries"            # Add task to inbox
  craft tasks add "Review PR" --schedule 2026-02-01
  craft tasks update ID --state done         # Mark task complete
  craft tasks delete ID                      # Delete task`,
}

var (
	taskScope      string
	taskDocumentID string
)

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks with optional filtering.

Scopes:
  all       - All task blocks in the space
  active    - All active (not done/canceled) tasks
  upcoming  - Tasks with upcoming schedule dates
  inbox     - Tasks in the inbox
  logbook   - Completed tasks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var tasks *models.TaskList

		if taskDocumentID != "" {
			tasks, err = client.GetDocumentTasks(taskDocumentID)
		} else {
			tasks, err = client.GetTasks(taskScope)
		}

		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == FormatJSON {
			return outputJSON(tasks)
		}
		return outputTasks(tasks.Items, format)
	},
}

var (
	taskMarkdown     string
	taskLocation     string
	taskScheduleDate string
	taskDeadlineDate string
	taskState        string
	taskJSON         string
	taskStdin        bool
)

var tasksAddCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Long: `Add a new task to your inbox or a document.

Locations:
  inbox    - Add to task inbox (default)
  document - Add to a specific document (requires --document)

Examples:
  craft tasks add "Buy groceries"
  craft tasks add "Review PR" --schedule 2026-02-01
  craft tasks add "Submit report" --deadline 2026-02-15
  craft tasks add "Meeting notes" --location document --document ID
  craft tasks add --json '{"tasks":[{"markdown":"Buy groceries","location":{"type":"inbox"}}]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksAddRaw(payload)
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		description := args[0]
		if taskLocation == "" {
			taskLocation = "inbox"
		}

		task, err := client.AddTask(description, taskLocation, taskDocumentID, taskScheduleDate, taskDeadlineDate)
		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(task.ID)
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(task)
		}
		fmt.Printf("Task created: %s (ID: %s)\n", task.Markdown, task.ID)
		return nil
	},
}

var tasksUpdateCmd = &cobra.Command{
	Use:   "update [task-id]",
	Short: "Update a task",
	Long: `Update a task's state or dates.

States:
  todo      - Mark as incomplete
  done      - Mark as complete
  canceled  - Mark as canceled

Examples:
  craft tasks update ID --state done
  craft tasks update ID --schedule 2026-02-01
  craft tasks update ID --deadline 2026-02-15
  craft tasks update --json '{"tasksToUpdate":[{"id":"ID","taskInfo":{"state":"done"}}]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksUpdateRaw(payload)
		}

		if taskState == "" && taskScheduleDate == "" && taskDeadlineDate == "" {
			return fmt.Errorf("at least one of --state, --schedule, or --deadline is required")
		}

		if isDryRun() {
			return dryRunOutput("update task", map[string]interface{}{"id": args[0]})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		taskID := args[0]
		if err := client.UpdateTask(taskID, taskState, taskScheduleDate, taskDeadlineDate); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Task %s updated\n", taskID)
		}
		return nil
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete [task-id]",
	Short: "Delete a task",
	Long: `Delete a task by its ID.

Examples:
  craft tasks delete ID
  craft tasks delete --json '{"idsToDelete":["ID"]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksDeleteRaw(payload)
		}

		if isDryRun() {
			return dryRunOutput("delete task", map[string]interface{}{
				"id": args[0], "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		taskID := args[0]
		if err := client.DeleteTask(taskID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Task %s deleted\n", taskID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)

	tasksCmd.AddCommand(tasksListCmd)
	tasksListCmd.Flags().StringVar(&taskScope, "scope", "all", "Filter scope: all, active, upcoming, inbox, logbook")
	tasksListCmd.Flags().StringVar(&taskDocumentID, "document", "", "Filter by document ID")

	tasksCmd.AddCommand(tasksAddCmd)
	tasksAddCmd.Flags().StringVar(&taskLocation, "location", "inbox", "Location: inbox, document")
	tasksAddCmd.Flags().StringVar(&taskDocumentID, "document", "", "Document ID (required for location=document)")
	tasksAddCmd.Flags().StringVar(&taskScheduleDate, "schedule", "", "Schedule date (YYYY-MM-DD)")
	tasksAddCmd.Flags().StringVar(&taskDeadlineDate, "deadline", "", "Deadline date (YYYY-MM-DD)")
	tasksAddCmd.Flags().StringVar(&taskJSON, "json", "", "Raw REST add tasks payload JSON")
	tasksAddCmd.Flags().BoolVar(&taskStdin, "stdin", false, "Read raw REST add tasks payload from stdin")

	tasksCmd.AddCommand(tasksUpdateCmd)
	tasksUpdateCmd.Flags().StringVar(&taskState, "state", "", "New state: todo, done, canceled")
	tasksUpdateCmd.Flags().StringVar(&taskScheduleDate, "schedule", "", "Schedule date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().StringVar(&taskDeadlineDate, "deadline", "", "Deadline date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().StringVar(&taskJSON, "json", "", "Raw REST update tasks payload JSON")
	tasksUpdateCmd.Flags().BoolVar(&taskStdin, "stdin", false, "Read raw REST update tasks payload from stdin")

	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksDeleteCmd.Flags().StringVar(&taskJSON, "json", "", "Raw REST delete tasks payload JSON")
	tasksDeleteCmd.Flags().BoolVar(&taskStdin, "stdin", false, "Read raw REST delete tasks payload from stdin")
}

func readTaskPayload(jsonPayload string, stdin bool) (map[string]interface{}, error) {
	if jsonPayload != "" && stdin {
		return nil, fmt.Errorf("--json and --stdin are mutually exclusive")
	}
	if stdin {
		input, err := readStdinString()
		if err != nil {
			return nil, err
		}
		return parseJSONObject(input)
	}
	return parseJSONObject(jsonPayload)
}

func runTasksAddRaw(payload map[string]interface{}) error {
	tasks, ok := payload["tasks"].([]interface{})
	if !ok || len(tasks) == 0 {
		return fmt.Errorf("raw tasks add payload must include non-empty \"tasks\" array")
	}
	if isDryRun() {
		return dryRunOutput("add tasks", map[string]interface{}{"payload": payload, "count": len(tasks)})
	}
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	result, err := client.AddTasksRaw(payload)
	if err != nil {
		return err
	}
	if isQuiet() {
		for _, task := range result {
			fmt.Println(task.ID)
		}
		return nil
	}
	return outputJSON(result)
}

func runTasksUpdateRaw(payload map[string]interface{}) error {
	tasks, ok := payload["tasksToUpdate"].([]interface{})
	if !ok || len(tasks) == 0 {
		return fmt.Errorf("raw tasks update payload must include non-empty \"tasksToUpdate\" array")
	}
	if isDryRun() {
		return dryRunOutput("update tasks", map[string]interface{}{"payload": payload, "count": len(tasks)})
	}
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	return client.UpdateTasksRaw(payload)
}

func runTasksDeleteRaw(payload map[string]interface{}) error {
	ids, ok := payload["idsToDelete"].([]interface{})
	if !ok || len(ids) == 0 {
		return fmt.Errorf("raw tasks delete payload must include non-empty \"idsToDelete\" array")
	}
	if isDryRun() {
		return dryRunOutput("delete tasks", map[string]interface{}{"payload": payload, "count": len(ids), "destructive": true})
	}
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	return client.DeleteTasksRaw(payload)
}

// outputTasks prints tasks in the specified format
func outputTasks(tasks []models.Task, format string) error {
	switch format {
	case FormatCompact:
		return outputJSON(tasks)
	case "table":
		return outputTasksTable(tasks)
	case "markdown":
		return outputTasksMarkdown(tasks)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputTasksTable prints tasks as a table
func outputTasksTable(tasks []models.Task) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tSTATE\tDESCRIPTION\tSCHEDULE\tDEADLINE")
		fmt.Fprintln(w, "---\t-----\t-----------\t--------\t--------")
	}

	for _, t := range tasks {
		desc := t.Markdown
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		schedule := t.ScheduleDate
		if schedule == "" {
			schedule = "-"
		}

		deadline := t.DeadlineDate
		if deadline == "" {
			deadline = "-"
		}

		stateIcon := "☐"
		switch t.State {
		case "done":
			stateIcon = "✅"
		case "canceled":
			stateIcon = "⊘"
		}

		fmt.Fprintf(w, "%s\t%s %s\t%s\t%s\t%s\n",
			t.ID, stateIcon, t.State, desc, schedule, deadline)
	}

	return w.Flush()
}

// outputTasksMarkdown prints tasks as markdown
func outputTasksMarkdown(tasks []models.Task) error {
	fmt.Println("# Tasks")
	for _, t := range tasks {
		checkbox := "[ ]"
		switch t.State {
		case "done":
			checkbox = "[x]"
		case "canceled":
			checkbox = "[-]"
		}

		fmt.Printf("- %s %s\n", checkbox, t.Markdown)

		if t.ScheduleDate != "" {
			fmt.Printf("  - **Scheduled**: %s\n", t.ScheduleDate)
		}
		if t.DeadlineDate != "" {
			fmt.Printf("  - **Deadline**: %s\n", t.DeadlineDate)
		}
		fmt.Printf("  - **ID**: %s\n", t.ID)
		fmt.Println()
	}
	return nil
}
