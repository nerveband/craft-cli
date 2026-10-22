package index

import (
	"github.com/ashrafali/craft-cli/internal/models"
	"os"
	"testing"
	"time"
)

func TestPrivateIndexSearchAndIsolation(t *testing.T) {
	dir := t.TempDir()
	path := Path(dir, "https://example.org/link", "fixture-key")
	if path == Path(dir, "https://example.org/link", "other-key") || path == Path(dir, "https://example.org/other-link", "fixture-key") {
		t.Fatal("connection identity collision")
	}
	s := New()
	s.Complete = true
	s.UpdatedAt = time.Now().UTC()
	s.Entries["a"] = Entry{Document: models.Document{ID: "a", Title: "Project planning", Markdown: "Review release notes"}}
	s.Entries["b"] = Entry{Document: models.Document{ID: "b", Title: "Shopping", Markdown: "Buy groceries"}}
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	hits := loaded.Search("RELEASE notes")
	if len(hits) != 1 || hits[0].DocumentID != "a" {
		t.Fatal(hits)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("cache permissions")
	}
	delete(loaded.Entries, "a")
	if err = loaded.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, _ = Load(path)
	if len(loaded.Search("release")) != 0 {
		t.Fatal("deleted posting remained")
	}
}
