package cmd

import (
	"strings"
	"testing"
)

func TestBuildSchemaIncludesBatchingAndFeatureFlags(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	del := findCommandSchema(rootSchema.Subcommands, "delete")
	if del == nil {
		t.Fatal("expected delete command in schema")
	}
	if !strings.Contains(del.Usage, "<document-id...>") {
		t.Errorf("delete usage %q does not advertise multiple ids", del.Usage)
	}

	mv := findCommandSchema(rootSchema.Subcommands, "move")
	if mv == nil {
		t.Fatal("expected move command in schema")
	}
	if !strings.Contains(mv.Usage, "<document-id...>") {
		t.Errorf("move usage %q does not advertise multiple ids", mv.Usage)
	}

	list := findCommandSchema(rootSchema.Subcommands, "list")
	if list == nil {
		t.Fatal("expected list command in schema")
	}
	for _, flag := range []string{"--daily-note-after", "--daily-note-before"} {
		if !hasFlagSchema(list.Flags, flag) {
			t.Errorf("list missing %s flag", flag)
		}
	}

	tasks := findCommandSchema(rootSchema.Subcommands, "tasks")
	if tasks == nil {
		t.Fatal("expected tasks command in schema")
	}
	tasksAdd := findCommandSchema(tasks.Subcommands, "add")
	if tasksAdd == nil {
		t.Fatal("expected tasks add in schema")
	}
	for _, flag := range []string{"--repeat", "--repeat-interval", "--repeat-weekdays",
		"--repeat-end", "--repeat-skip-weekends", "--repeat-dynamic-days",
		"--repeat-reminder", "--repeat-frequency", "--save-revert", "--diff"} {
		if !hasFlagSchema(tasksAdd.Flags, flag) {
			t.Errorf("tasks add missing %s flag", flag)
		}
	}

	upd := findCommandSchema(rootSchema.Subcommands, "update")
	if upd == nil {
		t.Fatal("expected update command in schema")
	}
	for _, flag := range []string{"--save-revert", "--diff"} {
		if !hasFlagSchema(upd.Flags, flag) {
			t.Errorf("update missing %s flag", flag)
		}
	}
}
