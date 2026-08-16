package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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
	taskMarkdown           string
	taskLocation           string
	taskScheduleDate       string
	taskDeadlineDate       string
	taskState              string
	taskJSON               string
	taskStdin              bool
	taskRepeatType         string
	taskRepeatFrequency    string
	taskRepeatInterval     int
	taskRepeatWeekdays     []int
	taskRepeatEnd          string
	taskRepeatReminder     string
	taskRepeatSkipWeekends bool
	taskRepeatDynamicDays  bool
	taskSaveRevert         string
	taskDiff               bool
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
  craft tasks add "Water plants" --repeat weekly --repeat-interval 2 --repeat-weekdays 1,3,5
  craft tasks add "Standup" --repeat daily --repeat-skip-weekends --repeat-reminder 09:00
  craft tasks add --json '{"tasks":[{"markdown":"Buy groceries","location":{"type":"inbox"}}]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rejectTaskRevertFlags(); err != nil {
			return err
		}
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

		repeat, err := buildRepeatConfig(repeatFlagsFromVars())
		if err != nil {
			return err
		}

		task, err := client.AddTask(description, taskLocation, taskDocumentID, taskScheduleDate, taskDeadlineDate, repeat)
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
		if err := rejectTaskRevertFlags(); err != nil {
			return err
		}
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksUpdateRaw(payload)
		}
		repeat, err := buildRepeatConfig(repeatFlagsFromVars())
		if err != nil {
			return err
		}
		if taskState == "" && taskScheduleDate == "" && taskDeadlineDate == "" && repeat == nil {
			return fmt.Errorf("at least one of --state, --schedule, --deadline, or a --repeat flag is required")
		}

		if isDryRun() {
			return dryRunOutput("update task", map[string]interface{}{"id": args[0]})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		taskID := args[0]
		if err := client.UpdateTask(taskID, taskState, taskScheduleDate, taskDeadlineDate, repeat); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Task %s updated\n", taskID)
		}
		return nil
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <task-id...>",
	Short: "Delete one or more tasks",
	Long: `Delete tasks by ID. Multiple IDs are sent in a single API request.

Examples:
  craft tasks delete ID
  craft tasks delete ID1 ID2 ID3
  craft tasks delete --json '{"idsToDelete":["ID"]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rejectTaskRevertFlags(); err != nil {
			return err
		}
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksDeleteRaw(payload)
		}
		if isDryRun() {
			if len(args) == 1 {
				return dryRunOutput("delete task", map[string]interface{}{
					"id": args[0], "destructive": true,
				})
			}
			return dryRunOutput("delete tasks", map[string]interface{}{
				"ids": args, "count": len(args), "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteTasks(args); err != nil {
			return err
		}

		if !isQuiet() {
			if len(args) == 1 {
				fmt.Printf("Task %s deleted\n", args[0])
			} else {
				fmt.Printf("%d tasks deleted\n", len(args))
			}
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

	tasksUpdateCmd.Flags().StringVar(&taskState, "state", "", "New state: todo, done, canceled")
	tasksUpdateCmd.Flags().StringVar(&taskScheduleDate, "schedule", "", "Schedule date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().StringVar(&taskDeadlineDate, "deadline", "", "Deadline date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().StringVar(&taskJSON, "json", "", "Raw REST update tasks payload JSON")
	tasksUpdateCmd.Flags().BoolVar(&taskStdin, "stdin", false, "Read raw REST update tasks payload from stdin")

	for _, c := range []*cobra.Command{tasksAddCmd, tasksUpdateCmd} {
		c.Flags().StringVar(&taskRepeatType, "repeat", "", "Repeat rule: daily, weekly, monthly, yearly")
		c.Flags().StringVar(&taskRepeatFrequency, "repeat-frequency", "", "Repeat frequency key (daily, weekly, monthly, yearly)")
		c.Flags().IntVar(&taskRepeatInterval, "repeat-interval", 0, "Repeat every N periods")
		c.Flags().IntSliceVar(&taskRepeatWeekdays, "repeat-weekdays", nil, "Repeat weekdays, 0=Sunday..6=Saturday (comma-separated)")
		c.Flags().StringVar(&taskRepeatEnd, "repeat-end", "", "Repeat end date (YYYY-MM-DD)")
		c.Flags().StringVar(&taskRepeatReminder, "repeat-reminder", "", "Reminder time (HH:MM)")
		c.Flags().BoolVar(&taskRepeatSkipWeekends, "repeat-skip-weekends", false, "Skip weekend occurrences")
		c.Flags().BoolVar(&taskRepeatDynamicDays, "repeat-dynamic-days", false, "Reschedule relative to completion date")
	}
	for _, c := range []*cobra.Command{tasksAddCmd, tasksUpdateCmd, tasksDeleteCmd} {
		c.Flags().StringVar(&taskSaveRevert, "save-revert", "", "Not supported for tasks; returns CAPABILITY_UNAVAILABLE (use craft blocks update --save-revert)")
		c.Flags().BoolVar(&taskDiff, "diff", false, "Not supported for tasks; returns CAPABILITY_UNAVAILABLE")
	}

	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksDeleteCmd.Flags().StringVar(&taskJSON, "json", "", "Raw REST delete tasks payload JSON")
	tasksDeleteCmd.Flags().BoolVar(&taskStdin, "stdin", false, "Read raw REST delete tasks payload from stdin")
}

// rejectTaskRevertFlags returns a structured error when --save-revert/--diff
// are used on tasks: Craft MCP exposes no task-write commands, so there is no
// revert metadata to capture for task mutations.
func rejectTaskRevertFlags() error {
	if taskSaveRevert == "" && !taskDiff {
		return nil
	}
	return newCLIError("CAPABILITY_UNAVAILABLE",
		"Craft MCP does not expose task writes; revert metadata is unavailable for tasks. For the task's underlying block, use craft blocks update <block-id> --save-revert")
}

// repeatFlagValues carries repeat flag inputs so validation is testable.
type repeatFlagValues struct {
	Type         string
	Frequency    string
	Interval     int
	Weekdays     []int
	EndDate      string
	Reminder     string
	SkipWeekends bool
	DynamicDays  bool
}

// buildRepeatConfig validates repeat flags and assembles a RepeatConfig.
// Returns (nil, nil) when no repeat flag was provided.
func buildRepeatConfig(v repeatFlagValues) (*models.RepeatConfig, error) {
	hasRule := v.Type != "" || v.Frequency != ""
	hasModifier := v.Interval != 0 || len(v.Weekdays) > 0 || v.EndDate != "" ||
		v.Reminder != "" || v.SkipWeekends || v.DynamicDays
	if !hasRule && !hasModifier {
		return nil, nil
	}
	if !hasRule {
		return nil, fmt.Errorf("repeat modifiers require --repeat or --repeat-frequency")
	}
	validRules := map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}
	if v.Type != "" && !validRules[v.Type] {
		return nil, fmt.Errorf("invalid --repeat %q (expected daily, weekly, monthly, or yearly)", v.Type)
	}
	if v.Frequency != "" && !validRules[v.Frequency] {
		return nil, fmt.Errorf("invalid --repeat-frequency %q (expected daily, weekly, monthly, or yearly)", v.Frequency)
	}
	if v.Interval < 0 {
		return nil, fmt.Errorf("--repeat-interval must be >= 1")
	}
	for _, day := range v.Weekdays {
		if day < 0 || day > 6 {
			return nil, fmt.Errorf("--repeat-weekdays values must be 0 (Sunday) through 6 (Saturday), got %d", day)
		}
	}
	if v.EndDate != "" {
		if _, err := time.Parse("2006-01-02", v.EndDate); err != nil {
			return nil, fmt.Errorf("invalid --repeat-end %q (expected YYYY-MM-DD)", v.EndDate)
		}
	}
	return &models.RepeatConfig{
		Type:         v.Type,
		Frequency:    v.Frequency,
		Interval:     v.Interval,
		Weekdays:     v.Weekdays,
		EndDate:      v.EndDate,
		Reminder:     v.Reminder,
		SkipWeekends: v.SkipWeekends,
		DynamicDays:  v.DynamicDays,
	}, nil
}

func repeatFlagsFromVars() repeatFlagValues {
	return repeatFlagValues{
		Type:         taskRepeatType,
		Frequency:    taskRepeatFrequency,
		Interval:     taskRepeatInterval,
		Weekdays:     taskRepeatWeekdays,
		EndDate:      taskRepeatEnd,
		Reminder:     taskRepeatReminder,
		SkipWeekends: taskRepeatSkipWeekends,
		DynamicDays:  taskRepeatDynamicDays,
	}
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
