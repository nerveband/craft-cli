package cmd

import "testing"

func TestTasksListDefaultScopeIsAll(t *testing.T) {
	flag := tasksListCmd.Flags().Lookup("scope")
	if flag == nil {
		t.Fatal("expected tasks list --scope flag")
	}
	if flag.DefValue != "all" {
		t.Fatalf("scope default = %q, want all", flag.DefValue)
	}
}
