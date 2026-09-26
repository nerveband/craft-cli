> Historical migration notes. The current source of truth is docs/contracts/craft-rest-space-openapi.json and the 2026-09-26 API audit plan. Do not implement old payload claims without checking that contract.

# Craft Connect API , Upgrade Map

**Date:** 2026-04-01
**craft-cli version:** v1.8.0
**API Docs:** https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1
**Test Endpoint:** https://connect.craft.do/links/HHRuPxZZTJ6/api/v1

---

## Coverage Matrix

### Fully Covered (21 stable endpoints)

| Resource | Endpoints | CLI Commands | Status |
|----------|-----------|-------------|--------|
| **Connection** | GET /connection | `craft connection`, `craft info` | Covered |
| **Documents** | GET /documents | `craft list` | Covered |
| | POST /documents | `craft create` | Covered |
| | DELETE /documents | `craft delete` | Covered |
| | PUT /documents/move | `craft move` | Covered |
| | GET /documents/search | `craft search` | Covered |
| **Blocks** | GET /blocks | `craft get`, `craft blocks get` | Covered |
| | POST /blocks | `craft blocks add` | Covered |
| | PUT /blocks | `craft blocks update` | Covered |
| | DELETE /blocks | `craft blocks delete` | Covered |
| | PUT /blocks/move | `craft blocks move` | Covered |
| | GET /blocks/search | `craft search` (block-level) | Covered |
| **Folders** | GET /folders | `craft folders list` | Covered |
| | POST /folders | `craft folders create` | Covered |
| | DELETE /folders | `craft folders delete` | Covered |
| | PUT /folders/move | `craft folders move` | Covered |
| **Tasks** | GET /tasks | `craft tasks` | Covered |
| | POST /tasks | `craft tasks add` | Covered |
| | PUT /tasks | `craft tasks update` | Covered |
| | DELETE /tasks | `craft tasks delete` | Covered |
| **Upload** | POST /upload | `craft upload` | Covered |

### Covered (Experimental, 9 endpoints)

| Resource | Endpoints | CLI Commands | Status |
|----------|-----------|-------------|--------|
| **Collections** | GET /collections | `craft collections` | Covered |
| | POST /collections | `craft collections create` | Covered |
| | GET /collections/{id}/schema | `craft collections schema` | Covered |
| | PUT /collections/{id}/schema | , | Needs verification |
| | GET /collections/{id}/items | `craft collections items` | Covered |
| | POST /collections/{id}/items | `craft collections add` | Covered |
| | PUT /collections/{id}/items | `craft collections update` | Covered |
| | DELETE /collections/{id}/items | `craft collections delete` | Covered |
| **Comments** | POST /comments | `craft comments add` | Covered |

### NOT Covered (5 endpoints) , NEW

| Resource | Endpoints | Proposed CLI Commands | Priority |
|----------|-----------|----------------------|----------|
| **Whiteboards** | POST /whiteboards | `craft whiteboards create` | Medium |
| | GET /whiteboards/{id}/elements | `craft whiteboards get` | Medium |
| | POST /whiteboards/{id}/elements | `craft whiteboards add` | Medium |
| | PUT /whiteboards/{id}/elements | `craft whiteboards update` | Medium |
| | DELETE /whiteboards/{id}/elements | `craft whiteboards delete` | Medium |

---

## New Features to Implement

### 1. Whiteboards (NEW , 5 endpoints)

Excalidraw-format whiteboard management. This is the only entirely missing API surface.

```
craft whiteboards create --page PAGE_ID              # Create empty whiteboard
craft whiteboards get WHITEBOARD_ID                   # Get elements (Excalidraw format)
craft whiteboards add WHITEBOARD_ID --json '[...]'    # Add elements
craft whiteboards update WHITEBOARD_ID --json '[...]' # Update elements
craft whiteboards delete WHITEBOARD_ID --ids "id1,id2" # Remove elements
```

### 2. Collections Schema Update (verify coverage)
- `PUT /collections/{id}/schema` , may need new `craft collections schema update` subcommand

### 3. API Enhancements to Existing Commands

| Enhancement | Affected Commands | Details |
|-------------|-------------------|---------|
| `fetchMetadata=true` | `craft list`, `craft get` | Add `--metadata` flag to include createdAt, lastModifiedAt, etc. |
| `Accept: text/markdown` | `craft get` | Add `--accept markdown` for native markdown from blocks endpoint |
| Date filters | `craft list` | Add `--from` and `--to` date range flags |
| RE2 regex search | `craft search` | Add `--regex` flag and `--case-sensitive` |
| Max depth control | `craft get`, `craft blocks get` | Add `--max-depth` flag |

---

## Testing Plan

Test against: `https://connect.craft.do/links/HHRuPxZZTJ6/api/v1` (no auth required)

### Phase 1: Verify Existing Coverage
```bash
./craft-cli list
./craft-cli folders
./craft-cli search "test"
./craft-cli get <doc-id>
./craft-cli blocks get <doc-id>
./craft-cli collections
./craft-cli tasks
./craft-cli connection
```

### Phase 2: Test New Whiteboard Endpoints
```bash
# Create a whiteboard (need a page ID first)
./craft-cli whiteboards create --page <page-id>
./craft-cli whiteboards get <whiteboard-id>
./craft-cli whiteboards add <whiteboard-id> --json '[{"type":"rectangle","x":0,"y":0,"width":100,"height":100}]'
./craft-cli whiteboards update <whiteboard-id> --json '[{"id":"elem-id","x":50}]'
./craft-cli whiteboards delete <whiteboard-id> --ids "elem-id"
```

### Phase 3: Test Enhanced Parameters
```bash
./craft-cli list --metadata --from 2026-01-01 --to 2026-04-01
./craft-cli get <doc-id> --max-depth 2
./craft-cli search "pattern" --regex --case-sensitive
```

---

## Priority Order

1. **Verify existing coverage** against live API , some endpoints may have drifted
2. **Whiteboards** , only entirely missing surface
3. **fetchMetadata flag** , quick win, adds useful info
4. **Date filters** , useful for list command
5. **Max depth control** , useful for large documents
6. **Collections schema update** , verify and add if missing
