# Craft CLI - MCP Server Feature Parity Design

**Date:** 2026-01-26
**Status:** ✅ Implemented
**Goal:** Achieve full feature parity between CLI and MCP server, supporting both LLM consumption and human-readable output

## Implementation Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Block model with all fields | ✅ Complete |
| Phase 1 | `--format structured` output | ✅ Complete |
| Phase 1 | `--format craft` output | ✅ Complete |
| Phase 2 | `--format rich` terminal output | ✅ Complete |
| Phase 3 | Folder commands (list/create/move/delete) | ✅ Complete |
| Phase 3 | Document move command | ✅ Complete |
| Phase 3 | List with --folder/--location filters | ✅ Complete |
| Phase 4 | Block operations (get/add/update/delete/move) | ✅ Complete |
| Phase 5 | Task management (list/add/update/delete) | ✅ Complete |

---

## Executive Summary

The Craft MCP server provides 28 tools with rich document structure preservation, while the CLI currently flattens content to plain markdown. This design outlines how to achieve complete parity for both machine-readable (LLM) and human-readable output.

---

## Gap Analysis

### Block Types & Attributes

| Feature | MCP Server | CLI Current | Gap |
|---------|------------|-------------|-----|
| **Block Types** |
| text | ✅ Full metadata | ✅ Markdown only | Missing attributes |
| page | ✅ With hierarchy | ❌ Flattened | **Critical** |
| table | ✅ Cell attributes | ✅ Markdown | Partial |
| code | ✅ With language | ✅ Basic | Missing language |
| line | ✅ lineStyle | ✅ Basic | Missing styles |
| image | ✅ url, altText | ✅ Markdown | Missing altText |
| file | ✅ url, fileName | ❌ Lost | **Missing** |
| richUrl | ✅ title, description, layout | ❌ Lost | **Missing** |
| **Styling** |
| textStyle | ✅ h1-h4, caption, card, page | Partial | Missing caption, card |
| listStyle | ✅ bullet, numbered, task, toggle | Partial | Missing toggle |
| decorations | ✅ callout, quote (combinable) | ❌ Raw tags | **Missing** |
| color | ✅ #RRGGBB hex | ❌ Lost | **Missing** |
| cardLayout | ✅ small, regular, large | ❌ Lost | **Missing** |
| indentationLevel | ✅ 0-5 | ❌ Lost | **Missing** |
| lineStyle | ✅ strong, regular, light, extraLight, pageBreak | ❌ Lost | **Missing** |
| font | ✅ system, serif, mono, rounded | ❌ Lost | **Missing** |
| textAlignment | ✅ left, center, right | ❌ Lost | **Missing** |
| **Tasks** |
| taskInfo.state | ✅ todo, done, canceled | ❌ Plain checkbox | **Critical** |
| taskInfo.completedAt | ✅ Timestamp | ❌ Lost | **Missing** |
| taskInfo.scheduleDate | ✅ Date | ❌ Lost | **Missing** |
| taskInfo.deadlineDate | ✅ Date | ❌ Lost | **Missing** |
| taskInfo.repeat | ✅ Full config | ❌ Lost | **Missing** |
| **Block Metadata** |
| Block IDs | ✅ All blocks | ❌ Lost | **Critical for editing** |
| comments | ✅ `<comment id="">` | ❌ Lost | Missing |

### Commands/Tools

| MCP Tool | CLI Command | Gap |
|----------|-------------|-----|
| documents_list | ✅ `craft list` | Partial - missing filters |
| documents_create | ✅ `craft create` | OK |
| documents_delete | ✅ `craft delete` | OK |
| documents_move | ❌ None | **Missing** |
| folders_list | ❌ None | **Missing** |
| folders_create | ❌ None | **Missing** |
| folders_move | ❌ None | **Missing** |
| folders_delete | ❌ None | **Missing** |
| blocks_get | ✅ `craft get` (partial) | Missing format options |
| blocks_add | ✅ `craft update` (partial) | Missing position control |
| blocks_update | ❌ None | **Missing** |
| blocks_delete | ❌ None | **Missing** |
| blocks_move | ❌ None | **Missing** |
| markdown_add | ❌ None | **Missing** |
| documents_search | ✅ `craft search` | OK |
| document_search | ❌ None | **Missing** |
| tasks_get | ❌ None | **Missing** |
| tasks_add | ❌ None | **Missing** |
| tasks_update | ❌ None | **Missing** |
| tasks_delete | ❌ None | **Missing** |
| collections_* (6 tools) | ❌ None | **Missing** |
| comments_add | ❌ None | **Missing** |
| connection_time_get | ❌ None | **Missing** |

---

## Design: Output Formats

### 1. Structured JSON (for LLMs)

New `--format structured` that returns full block tree:

```json
{
  "id": "4c6d40b0-7d98-81bb-b7af-ab002911a64d",
  "type": "page",
  "textStyle": "page",
  "title": "ISLA Ramadan 2026 Campaign Playbook",
  "content": [
    {
      "id": "281e065f-3055-f4cc-5797-eeb48e0b6cd5",
      "type": "text",
      "markdown": "**Goal: $100,000 | Timeline:** Feb 17 - Mar 30, 2026"
    },
    {
      "id": "eb15e62b-5657-2f7d-b77c-d664c474fb83",
      "type": "page",
      "textStyle": "card",
      "cardLayout": "regular",
      "title": "Quick Reference",
      "content": [...]
    }
  ]
}
```

### 2. Craft Markdown (matching MCP)

New `--format craft` that outputs the MCP-style markdown with XML tags:

```markdown
<page id="4c6d40b0-7d98-81bb-b7af-ab002911a64d">
  <pageTitle>ISLA Ramadan 2026 Campaign Playbook</pageTitle>
  <content>
    **Goal: $100,000 | Timeline:** Feb 17 - Mar 30, 2026

    <page id="eb15e62b-5657-2f7d-b77c-d664c474fb83" textStyle="card" cardLayout="regular">
      <pageTitle>Quick Reference</pageTitle>
      <content>
        | **Element** | **Details** |
        ...
      </content>
    </page>
  </content>
</page>
```

### 3. Rich Terminal Output (for humans)

New `--format rich` with ANSI colors and Unicode:

```
╔══════════════════════════════════════════════════════════════╗
║  ISLA Ramadan 2026 Campaign Playbook                         ║
╚══════════════════════════════════════════════════════════════╝

**Goal: $100,000 | Timeline:** Feb 17 - Mar 30, 2026

┌─ 📋 Quick Reference ─────────────────────────────────────────┐
│ │ Element          │ Details                                │
│ │──────────────────│────────────────────────────────────────│
│ │ Tagline          │ Your Gift Strengthens Every Islamic... │
│ │ Social Hashtag   │ #BeTheRipple                           │
└──────────────────────────────────────────────────────────────┘

┌─ 📋 The Core Pitch ──────────────────────────────────────────┐
│ **Why ISLA in addition to my local school?**                  │
│                                                               │
│ Giving to ISLA does not compete with giving to your school.  │
│ ░░ It multiplies what your school can accomplish. ░░         │
│ (highlighted in blue)                                         │
└──────────────────────────────────────────────────────────────┘

Tasks:
  ✅ Set up Main Codebase (completed 2023-05-01)
  ✅ Set up all branches (completed 2023-05-01)
  ☐ Set up unit-testing for all dev-based features
    ☐ Connect CMS to Database, blog

🟠 Note that this is just a rough estimate...
   (callout block - orange)
```

---

## Data Model Changes

### Enhanced Block Model

```go
// Block represents a content block from the Craft blocks API
type Block struct {
    ID               string        `json:"id"`
    Type             string        `json:"type"`
    TextStyle        string        `json:"textStyle,omitempty"`
    Markdown         string        `json:"markdown,omitempty"`
    Content          []Block       `json:"content,omitempty"`

    // New fields for parity
    ListStyle        string        `json:"listStyle,omitempty"`        // bullet, numbered, task, toggle
    Decorations      []string      `json:"decorations,omitempty"`      // callout, quote
    Color            string        `json:"color,omitempty"`            // #RRGGBB
    CardLayout       string        `json:"cardLayout,omitempty"`       // small, square, regular, large
    IndentationLevel int           `json:"indentationLevel,omitempty"` // 0-5
    LineStyle        string        `json:"lineStyle,omitempty"`        // strong, regular, light, etc.
    Font             string        `json:"font,omitempty"`             // system, serif, mono, rounded
    TextAlignment    string        `json:"textAlignment,omitempty"`    // left, center, right

    // Task-specific
    TaskInfo         *TaskInfo     `json:"taskInfo,omitempty"`

    // Media blocks
    URL              string        `json:"url,omitempty"`
    AltText          string        `json:"altText,omitempty"`
    FileName         string        `json:"fileName,omitempty"`

    // Rich URL
    Title            string        `json:"title,omitempty"`
    Description      string        `json:"description,omitempty"`
    Layout           string        `json:"layout,omitempty"`

    // Code blocks
    Language         string        `json:"language,omitempty"`
    RawCode          string        `json:"rawCode,omitempty"`

    // Table
    Rows             [][]TableCell `json:"rows,omitempty"`
}

type TaskInfo struct {
    State         string     `json:"state,omitempty"`         // todo, done, canceled
    CompletedAt   *time.Time `json:"completedAt,omitempty"`
    CanceledAt    *time.Time `json:"canceledAt,omitempty"`
    ScheduleDate  string     `json:"scheduleDate,omitempty"`  // YYYY-MM-DD
    DeadlineDate  string     `json:"deadlineDate,omitempty"`  // YYYY-MM-DD
    Repeat        *RepeatConfig `json:"repeat,omitempty"`
}

type TableCell struct {
    Value      string      `json:"value"`
    Attributes []TextAttr  `json:"attributes,omitempty"`
}

type TextAttr struct {
    Type  string `json:"type"`  // bold, italic, highlight, etc.
    Start int    `json:"start"`
    End   int    `json:"end"`
    Color string `json:"color,omitempty"` // for highlights
}
```

---

## New Commands

### Folder Management

```bash
# List all folders with document counts
craft folders list [--format json|table|tree]

# Create folder
craft folders create "New Folder" [--parent <folder-id>]

# Move folder
craft folders move <folder-id> --to <parent-folder-id>
craft folders move <folder-id> --to root

# Delete folder (contents move to parent)
craft folders delete <folder-id>
```

### Document Management Enhancements

```bash
# List with folder filter
craft list --folder <folder-id>
craft list --location unsorted|trash|templates|daily_notes

# Move document
craft move <doc-id> --to <folder-id>
craft move <doc-id> --to unsorted

# Get with format options
craft get <doc-id> --format structured   # Full JSON tree
craft get <doc-id> --format craft        # MCP-style XML markdown
craft get <doc-id> --format rich         # Terminal with colors
craft get <doc-id> --format markdown     # Plain markdown (current default)
craft get <doc-id> --format json         # Full block response (current)
```

### Block Operations

```bash
# Get specific block
craft blocks get <block-id> [--depth <n>]

# Add block at position
craft blocks add <page-id> --markdown "..." --position start|end
craft blocks add --sibling <block-id> --position before|after --markdown "..."

# Update block
craft blocks update <block-id> --markdown "new content"

# Delete block
craft blocks delete <block-id>

# Move block
craft blocks move <block-id> --to <page-id> --position start|end
```

### Task Management

```bash
# List tasks
craft tasks list --scope active|upcoming|inbox|logbook
craft tasks list --document <doc-id>

# Add task
craft tasks add "Task description" --location inbox
craft tasks add "Task description" --date today --schedule 2026-02-01

# Update task
craft tasks update <task-id> --state done
craft tasks update <task-id> --schedule 2026-02-15 --deadline 2026-02-20

# Delete task
craft tasks delete <task-id>
```

### Search Enhancements

```bash
# Search within document
craft search --document <doc-id> "pattern" [--regex] [--context 5]

# Search with date filters
craft search "query" --created-after 2026-01-01 --modified-before 2026-01-15
```

---

## Implementation Phases

### Phase 1: Core Output Parity (Week 1)
1. Update Block model with all fields
2. Implement `--format structured` output
3. Implement `--format craft` output (MCP-style markdown)
4. Preserve block IDs in all outputs
5. Handle all block types properly

### Phase 2: Rich Terminal Output (Week 2)
1. Implement `--format rich` with ANSI colors
2. Card/sub-page visual boxes
3. Task checkboxes with state colors
4. Callout rendering with background colors
5. Indentation preservation

### Phase 3: New Commands - Folders (Week 3)
1. `craft folders list|create|move|delete`
2. `craft list --folder` filter
3. `craft move` command for documents

### Phase 4: Block Operations (Week 4)
1. `craft blocks get|add|update|delete|move`
2. Position-aware block insertion
3. Markdown to blocks conversion

### Phase 5: Task Management (Week 5)
1. `craft tasks list|add|update|delete`
2. Task state management
3. Scheduling and repeat support

### Phase 6: Advanced Features (Week 6+)
1. Collections support
2. Comments support
3. Document search within block
4. Daily notes support

---

## Success Criteria

1. **LLM Consumption**: `craft get --format structured` returns identical data structure to MCP `blocks_get` with `format=json`

2. **Human Readability**: `craft get --format rich` displays documents with visual hierarchy, colors, and proper formatting

3. **Feature Coverage**: All 28 MCP tools have CLI equivalents

4. **Performance**: CLI should be faster than MCP due to direct API access (no JSON-RPC overhead)

---

## Testing Strategy

1. **Comparison Tests**: For each document, compare MCP output vs CLI output
2. **Round-trip Tests**: Create document via CLI, verify via MCP, and vice versa
3. **Visual Regression**: Screenshot tests for `--format rich` output
4. **Sample Documents**: Use 10+ diverse documents covering all block types

### Test Documents (from analysis)

| Document | Features to Test |
|----------|------------------|
| ISLA Playbook | Sub-pages (9), cards (5 with 3 layouts: small/regular/large), tables, tasks (31), fonts (serif/mono/rounded), toggle lists, callouts with color (#0064ff), quote+callout combo, caption text, code blocks, indentation levels |
| Sprint Backlog | Tasks with states, callouts, colors, dates |
| IG Moodboard | Images, cards, highlights, comments |
| Contract | File attachments |
| Email Swipe File | Files, basic text |

---

## API Reference

### MCP Server Tools → CLI Command Mapping

| MCP Tool | CLI Command | Notes |
|----------|-------------|-------|
| documents_list | `craft list` | Add --folder, --location |
| documents_create | `craft create` | OK |
| documents_move | `craft move` | New |
| documents_delete | `craft delete` | OK |
| folders_list | `craft folders list` | New |
| folders_create | `craft folders create` | New |
| folders_move | `craft folders move` | New |
| folders_delete | `craft folders delete` | New |
| blocks_get | `craft get --format structured` | Enhanced |
| blocks_add | `craft blocks add` | New |
| blocks_update | `craft blocks update` | New |
| blocks_delete | `craft blocks delete` | New |
| blocks_move | `craft blocks move` | New |
| markdown_add | `craft blocks add --markdown` | Combined |
| documents_search | `craft search` | OK |
| document_search | `craft search --document` | Enhanced |
| tasks_get | `craft tasks list` | New |
| tasks_add | `craft tasks add` | New |
| tasks_update | `craft tasks update` | New |
| tasks_delete | `craft tasks delete` | New |
| collections_* | `craft collections *` | New (future) |
| comments_add | `craft comments add` | New (future) |
| connection_time_get | `craft info --time` | New |

---

## API Limitations (Discovered)

**Document-level styling NOT exposed via Craft Connect API:**

These properties exist in Craft but cannot be read or modified via MCP or CLI:
- Backdrop/background image
- Document background color
- Cover image
- Default document font
- Wide page setting
- Separator style

**Block-level styling IS available (verified in ISLA document):**

| Property | Values Found | Count |
|----------|--------------|-------|
| `cardLayout` | small, regular, large | 5 blocks |
| `font` | system (default), serif, mono, rounded | 16 blocks |
| `decorations` | quote, callout (can combine) | 2 blocks |
| `color` | #RRGGBB hex (e.g., #0064ff) | 1 block |
| `lineStyle` | extraLight | 77 blocks |
| `listStyle` | bullet, numbered, task, toggle | 296 blocks |
| `indentationLevel` | 1, 2 | 3 blocks |
| `taskInfo` | state: todo/done/canceled | 31 blocks |
| `textStyle` | h1, h2, h3, page, card, caption | 108 blocks |

---

## Open Questions

1. **Default format**: Should `craft get` default to `structured` (for LLMs) or `markdown` (current behavior)?
2. **Block ID visibility**: Should block IDs always be shown, or only with `--verbose`?
3. **Color scheme**: What color palette for `--format rich` terminal output?
