package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ashrafali/craft-cli/internal/api"
	localindex "github.com/ashrafali/craft-cli/internal/index"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
	"path/filepath"
	"sort"
	"time"
)

var syncFull bool
var cacheMaxAge time.Duration
var responseSource = "live"
var responseSnapshot time.Time

func indexPath() (string, error) {
	endpoint, key, _, err := resolveREST()
	if err != nil {
		return "", err
	}
	configPath, _ := cfgManager.Paths()
	return localindex.Path(filepath.Dir(configPath), endpoint, key), nil
}
func loadLocalSnapshot() (*localindex.Snapshot, error) {
	path, err := indexPath()
	if err != nil {
		return nil, err
	}
	s, err := localindex.Load(path)
	if err != nil {
		return nil, err
	}
	if !s.Complete {
		return nil, fmt.Errorf("local index is incomplete; rerun craft sync")
	}
	responseSource = "local"
	responseBackend = "local-index"
	responseSnapshot = s.UpdatedAt
	return s, nil
}
func useLocalData() (bool, error) {
	if dataSource == "local" {
		if outputFormat == "raw" {
			return false, fmt.Errorf("raw upstream bytes require --data-source live; use json for local snapshots")
		}
		return true, nil
	}
	if outputFormat == "raw" {
		return false, nil
	}
	if dataSource != "auto" {
		return false, nil
	}
	path, err := indexPath()
	if err != nil {
		return false, err
	}
	s, err := localindex.Load(path)
	return err == nil && s.Complete && time.Since(s.UpdatedAt) <= cacheMaxAge, nil
}

var syncCmd = &cobra.Command{Use: "sync", Short: "Incrementally index documents for private offline list, get and search", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	path, err := indexPath()
	if err != nil {
		return err
	}
	snapshot, err := localindex.Load(path)
	if err != nil || syncFull {
		snapshot = localindex.New()
	}
	docs, err := client.GetDocumentsAdvanced(api.ListDocumentsOptions{FetchMetadata: true})
	if err != nil {
		return err
	}
	wanted := map[string]bool{}
	updated := 0
	for _, doc := range docs.Items {
		wanted[doc.ID] = true
	}
	for id := range snapshot.Entries {
		if !wanted[id] {
			delete(snapshot.Entries, id)
		}
	}
	snapshot.Complete = false
	// Persist completed reads so an interrupted or rate-limited refresh resumes.
	for _, doc := range docs.Items {
		old, ok := snapshot.Entries[doc.ID]
		if ok && !doc.LastModifiedAt.IsZero() && old.Modified.Equal(doc.LastModifiedAt) {
			continue
		}
		blocks, err := client.GetDocumentBlocksWithDepth(doc.ID, -1)
		if err != nil {
			return err
		}
		doc.Markdown = api.CombineBlocksMarkdown(blocks, true)
		snapshot.Entries[doc.ID] = localindex.Entry{Document: doc, Blocks: blocks, Modified: doc.LastModifiedAt}
		if err = snapshot.Save(path); err != nil {
			return err
		}
		updated++
		time.Sleep(650 * time.Millisecond)
	}
	snapshot.Complete = true
	snapshot.UpdatedAt = time.Now().UTC()
	if err = snapshot.Save(path); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"documents": len(snapshot.Entries), "updated": updated, "updated_at": snapshot.UpdatedAt, "status": "synced"})
}}

func localList() error {
	if listFolderID != "" || listLocation != "" || listCreatedAfter != "" || listCreatedBefore != "" || listModifiedAfter != "" || listModifiedBefore != "" || listDailyNoteAfter != "" || listDailyNoteBefore != "" || listCursor != "" || listOffset != 0 {
		return fmt.Errorf("local list does not support location/date/folder filters or cursors; use --data-source live")
	}
	s, err := loadLocalSnapshot()
	if err != nil {
		return err
	}
	docs := []models.Document{}
	for _, entry := range s.Entries {
		doc := entry.Document
		doc.Markdown = ""
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	total := len(docs)
	if listCount {
		return outputCount(total)
	}
	if listLimit > 0 && len(docs) > listLimit {
		docs = docs[:listLimit]
	}
	if listFields != "" {
		items, err := projectItems(docs, listFields)
		if err != nil {
			return err
		}
		return outputJSON(payloadWithMetadata(items, total, len(docs) < total, total, len(docs), "craft list --data-source local --limit 0"))
	}
	if isJSONFormat(getOutputFormat()) {
		return outputJSON(payloadWithMetadata(docs, total, len(docs) < total, total, len(docs), "craft list --data-source local --limit 0"))
	}
	return outputDocuments(docs, getOutputFormat())
}
func localGet(id string) error {
	s, err := loadLocalSnapshot()
	if err != nil {
		return err
	}
	entry, ok := s.Entries[id]
	if !ok {
		return newCLIError("NOT_FOUND", "document not present in local index; run craft sync")
	}
	// Clone through JSON so depth trimming cannot mutate shared cached state.
	data, _ := json.Marshal(entry.Blocks)
	var blocks models.BlocksResponse
	if err = json.Unmarshal(data, &blocks); err != nil {
		return err
	}
	var trim func([]models.Block, int) []models.Block
	trim = func(items []models.Block, depth int) []models.Block {
		if depth == 0 {
			return nil
		}
		for i := range items {
			items[i].Content = trim(items[i].Content, depth-1)
			delete(items[i].Extra, "content")
		}
		return items
	}
	if getMaxDepth >= 0 {
		blocks.Content = trim(blocks.Content, getMaxDepth)
		delete(blocks.Extra, "content")
	}
	format := getOutputFormat()
	if outputFile != "" {
		if format == FormatStructured || format == FormatCraft || format == FormatRich {
			return writeBlocksToFile(&blocks, format)
		}
		return writeAtomicOutput(outputFile, []byte(api.CombineBlocksMarkdown(blocks, true)))
	}
	switch format {
	case FormatStructured:
		return outputBlocksStructured(&blocks)
	case FormatCraft:
		return outputBlocksCraft(&blocks)
	case FormatRich:
		return outputBlocksRich(&blocks)
	}
	doc := entry.Document
	doc.Markdown = api.CombineBlocksMarkdown(blocks, true)
	return outputDocument(&doc, format)
}
func localSearch(args []string) error {
	if len(args) != 1 || searchDate != "" || searchSelectedDocument != "" || searchRegex != "" || len(searchFolders) > 0 || len(searchDocuments) > 0 || searchLocation != "" || searchCreatedAfter != "" || searchCreatedBefore != "" || searchModifiedAfter != "" || searchModifiedBefore != "" || searchDailyAfter != "" || searchDailyBefore != "" || searchMetadata || searchCaseSensitive {
		return fmt.Errorf("local search accepts one full-text query without REST scope/regex/context flags; use --data-source live for those filters")
	}
	s, err := loadLocalSnapshot()
	if err != nil {
		return err
	}
	items := s.Search(args[0])
	total := len(items)
	if searchCount {
		return outputCount(total)
	}
	if searchLimit > 0 && len(items) > searchLimit {
		items = items[:searchLimit]
	}
	if searchFields != "" {
		projected, err := projectItems(items, searchFields)
		if err != nil {
			return err
		}
		return outputJSON(payloadWithMetadata(projected, total, len(items) < total, total, len(items), "craft search --data-source local --limit 0"))
	}
	return outputJSON(payloadWithMetadata(items, total, len(items) < total, total, len(items), "craft search --data-source local --limit 0"))
}
func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncFull, "full", false, "Refetch every document rather than only changed metadata")
	rootCmd.PersistentFlags().DurationVar(&cacheMaxAge, "cache-max-age", 5*time.Minute, "Maximum index age for --data-source auto; local explicitly permits older snapshots")
	commandEffects["sync"] = CommandEffect{"non_idempotent", "local", false}
}
