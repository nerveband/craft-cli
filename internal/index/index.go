// Package index stores opt-in, private snapshots partitioned by connection identity.
package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ashrafali/craft-cli/internal/models"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Entry struct {
	Document models.Document       `json:"document"`
	Blocks   models.BlocksResponse `json:"blocks"`
	Modified time.Time             `json:"modified"`
}
type Snapshot struct {
	Version   int                 `json:"version"`
	UpdatedAt time.Time           `json:"updated_at"`
	Complete  bool                `json:"complete"`
	Entries   map[string]Entry    `json:"entries"`
	Terms     map[string][]string `json:"terms"`
}

func Path(dir, endpoint, key string) string {
	sum := sha256.Sum256([]byte(endpoint + "\x00" + key))
	return filepath.Join(dir, "index", hex.EncodeToString(sum[:])+".json")
}
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("local index unavailable; run craft sync: %w", err)
	}
	var value Snapshot
	if err = json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("invalid local index: %w", err)
	}
	if value.Version != 1 || value.Entries == nil {
		return nil, fmt.Errorf("unsupported local index; run craft sync --full")
	}
	return &value, nil
}
func New() *Snapshot {
	return &Snapshot{Version: 1, Entries: map[string]Entry{}, Terms: map[string][]string{}}
}
func tokens(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
}
func (s *Snapshot) Save(path string) error {
	s.Terms = map[string][]string{}
	ids := make([]string, 0, len(s.Entries))
	for id := range s.Entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		entry := s.Entries[id]
		seen := map[string]bool{}
		for _, term := range tokens(entry.Document.Title + " " + entry.Document.Markdown) {
			if !seen[term] {
				s.Terms[term] = append(s.Terms[term], id)
				seen[term] = true
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".index-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = json.NewEncoder(f).Encode(s); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func (s *Snapshot) Search(query string) []models.SearchItem {
	terms := tokens(query)
	counts := map[string]int{}
	for _, term := range terms {
		for _, id := range s.Terms[term] {
			counts[id]++
		}
	}
	ids := []string{}
	for id, count := range counts {
		if count == len(terms) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	result := []models.SearchItem{}
	for _, id := range ids {
		entry := s.Entries[id]
		snippet := entry.Document.Markdown
		if len(snippet) > 600 {
			snippet = string([]rune(snippet)[:min(600, len([]rune(snippet)))])
		}
		result = append(result, models.SearchItem{DocumentID: id, Markdown: snippet})
	}
	return result
}
