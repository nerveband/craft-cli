package cmd

import "testing"

func TestParseMCPBatchInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantTool  string
	}{
		{
			name:      "newline commands",
			input:     "connection info\n# comment\nblocks explore-themes\n",
			wantCount: 2,
			wantTool:  "craft_read",
		},
		{
			name:      "json string array",
			input:     `["connection info","blocks explore-washi"]`,
			wantCount: 2,
			wantTool:  "craft_read",
		},
		{
			name:      "json object array",
			input:     `[{"tool":"craft_write","command":"documents create --title Test"}]`,
			wantCount: 1,
			wantTool:  "craft_write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := parseMCPBatchInput(tt.input, "craft_read")
			if err != nil {
				t.Fatalf("parseMCPBatchInput() error = %v", err)
			}
			if len(ops) != tt.wantCount {
				t.Fatalf("len(ops) = %d, want %d", len(ops), tt.wantCount)
			}
			if ops[0].Tool != tt.wantTool {
				t.Fatalf("ops[0].Tool = %q, want %q", ops[0].Tool, tt.wantTool)
			}
		})
	}
}

func TestReadRevertInfoPayloadWrapsRawObject(t *testing.T) {
	oldInfo := blockRevertInfo
	oldFile := blockRevertInfoFile
	t.Cleanup(func() {
		blockRevertInfo = oldInfo
		blockRevertInfoFile = oldFile
	})

	blockRevertInfo = `{"operation":"add","blockIds":["b1"],"expectedStamps":{}}`
	blockRevertInfoFile = ""

	payload, err := readRevertInfoPayload()
	if err != nil {
		t.Fatalf("readRevertInfoPayload() error = %v", err)
	}
	if _, ok := payload["revertInfo"]; !ok {
		t.Fatalf("expected revertInfo wrapper, got %#v", payload)
	}
}

func TestBuildMCPBlocksCommandIncludesStyleFlags(t *testing.T) {
	command, err := buildMCPBlocksCommand("update", []map[string]interface{}{{
		"id":       "block123",
		"themeId":  "fire-horse",
		"coverURL": "https://example.com/cover.jpg",
	}}, nil)
	if err != nil {
		t.Fatalf("buildMCPBlocksCommand() error = %v", err)
	}
	want := `blocks update --id "block123" --theme-id "fire-horse" --cover-url "https://example.com/cover.jpg"`
	if command != want {
		t.Fatalf("command = %q, want %q", command, want)
	}
}

func TestFindRevertInfoInsideMCPTextPayload(t *testing.T) {
	payload := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"text": `{"ok":true,"revertInfo":{"operation":"update","blockIds":["b1"]}}`,
			},
		},
	}
	info, ok := findRevertInfo(payload)
	if !ok {
		t.Fatal("expected revertInfo")
	}
	infoMap, ok := info.(map[string]interface{})
	if !ok {
		t.Fatalf("revertInfo type = %T, want map", info)
	}
	if infoMap["operation"] != "update" {
		t.Fatalf("operation = %v, want update", infoMap["operation"])
	}
}

func TestPropertyFlagToMCP(t *testing.T) {
	got := propertyFlagToMCP("Status=select")
	want := `--Status "select"`
	if got != want {
		t.Fatalf("propertyFlagToMCP() = %q, want %q", got, want)
	}

	got = propertyFlagToMCP("Due Date=2026-07-09")
	want = `--"Due Date" "2026-07-09"`
	if got != want {
		t.Fatalf("propertyFlagToMCP() = %q, want %q", got, want)
	}
}

func TestExtractMarkdownImageURLs(t *testing.T) {
	command := `blocks add --page "page1" --markdown "Intro ![Logo](https://r.craft.do/JZV3hKXgrU-zmCN9cRVnn) and ![Public](https://example.com/a.png "caption")"`
	urls := extractMarkdownImageURLs(command)
	if len(urls) != 2 {
		t.Fatalf("len(urls) = %d, want 2: %#v", len(urls), urls)
	}
	if urls[0] != "https://r.craft.do/JZV3hKXgrU-zmCN9cRVnn" {
		t.Fatalf("urls[0] = %q", urls[0])
	}
	if urls[1] != "https://example.com/a.png" {
		t.Fatalf("urls[1] = %q", urls[1])
	}
}

func TestEnhanceMCPWriteErrorForMarkdownImageNotFound(t *testing.T) {
	err := enhanceMCPWriteError(
		`blocks add --page "page1" --markdown "![Logo](https://r.craft.do/dead)"`,
		newCLIError("NOT_FOUND", "Document not found"),
	)
	coded, ok := err.(*cliCodeError)
	if !ok {
		t.Fatalf("error type = %T, want *cliCodeError", err)
	}
	if coded.Code != "IMAGE_ASSET_UNAVAILABLE" {
		t.Fatalf("code = %q, want IMAGE_ASSET_UNAVAILABLE", coded.Code)
	}
	if !contains(coded.Message, "https://r.craft.do/dead") {
		t.Fatalf("message does not include image URL: %q", coded.Message)
	}
}

func TestEnhanceMCPWriteErrorLeavesPlainNotFoundAlone(t *testing.T) {
	original := newCLIError("NOT_FOUND", "Document not found")
	err := enhanceMCPWriteError(`blocks add --page "missing" --markdown "Plain text"`, original)
	if err != original {
		t.Fatalf("expected original error, got %#v", err)
	}
}
