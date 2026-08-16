package cmd

import "testing"

func TestResolveSearchScope(t *testing.T) {
	t.Run("single document means block search", func(t *testing.T) {
		scope := resolveSearchScope(nil, []string{"doc1"})
		if scope.BlockDocID != "doc1" {
			t.Errorf("BlockDocID = %q, want doc1", scope.BlockDocID)
		}
		if scope.DocumentIDs != "" {
			t.Errorf("DocumentIDs = %q, want empty", scope.DocumentIDs)
		}
	})

	t.Run("multiple documents mean scoped document search", func(t *testing.T) {
		scope := resolveSearchScope(nil, []string{"doc1", "doc2"})
		if scope.BlockDocID != "" {
			t.Errorf("BlockDocID = %q, want empty", scope.BlockDocID)
		}
		if scope.DocumentIDs != "doc1,doc2" {
			t.Errorf("DocumentIDs = %q, want doc1,doc2", scope.DocumentIDs)
		}
	})

	t.Run("folders join comma-separated", func(t *testing.T) {
		scope := resolveSearchScope([]string{"f1", "f2", "f3"}, nil)
		if scope.FolderIDs != "f1,f2,f3" {
			t.Errorf("FolderIDs = %q, want f1,f2,f3", scope.FolderIDs)
		}
	})

	t.Run("empty inputs produce empty scope", func(t *testing.T) {
		scope := resolveSearchScope(nil, nil)
		if scope.BlockDocID != "" || scope.FolderIDs != "" || scope.DocumentIDs != "" {
			t.Errorf("expected empty scope, got %+v", scope)
		}
	})
}
