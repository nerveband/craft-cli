package cmd

import "testing"

func TestDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := deleteCmd.Args(deleteCmd, []string{"doc1", "doc2", "doc3"}); err != nil {
		t.Fatalf("deleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := deleteCmd.Args(deleteCmd, []string{}); err == nil {
		t.Fatal("deleteCmd.Args accepted zero ids")
	}
}

func TestBlocksDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := blocksDeleteCmd.Args(blocksDeleteCmd, []string{"b1", "b2"}); err != nil {
		t.Fatalf("blocksDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := blocksDeleteCmd.Args(blocksDeleteCmd, []string{}); err == nil {
		t.Fatal("blocksDeleteCmd.Args accepted zero ids")
	}
}

func TestTasksDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := tasksDeleteCmd.Args(tasksDeleteCmd, []string{"t1", "t2"}); err != nil {
		t.Fatalf("tasksDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := tasksDeleteCmd.Args(tasksDeleteCmd, []string{}); err == nil {
		t.Fatal("tasksDeleteCmd.Args accepted zero ids")
	}
}

func TestFoldersDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := foldersDeleteCmd.Args(foldersDeleteCmd, []string{"f1", "f2"}); err != nil {
		t.Fatalf("foldersDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := foldersDeleteCmd.Args(foldersDeleteCmd, []string{}); err == nil {
		t.Fatal("foldersDeleteCmd.Args accepted zero ids")
	}
}

func TestMoveCmdAcceptsMultipleIDs(t *testing.T) {
	if err := moveCmd.Args(moveCmd, []string{"doc1", "doc2"}); err != nil {
		t.Fatalf("moveCmd.Args rejected multiple ids: %v", err)
	}
	if err := moveCmd.Args(moveCmd, []string{}); err == nil {
		t.Fatal("moveCmd.Args accepted zero ids")
	}
}
