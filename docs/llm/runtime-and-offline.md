# Runtime contracts and offline data in 3.0

`craft schema --command 'tasks update'` reads no configuration and makes no requests. It describes registered commands, effects, typed stdout schemas, flag and argument contracts, examples, and output variants. Shared definitions use JSON pointers within the manifest. Use `--cli-request-schema` or `--cli-response-schema` on any command for a standalone CLI contract. Existing `--request-schema` and `--response-schema` describe upstream REST payloads, which can differ from CLI convenience output.

MCP tools define their own variable results. The CLI preserves those results and declares their dynamic boundary rather than inventing fields. Read `craft mcp tools` for the selected server. Human text, raw bytes, counts, projections, compact arrays and dry runs have separate output variants. Normal JSON objects include `_provenance` with scope, source and backend. Local index reads also report snapshot time and age. Raw responses, scalar projections and shell completion remain unchanged streams.

## Upgrade from 2.x

Document lists now default to 50 records. Document and block reads default to depth 2. Collection items default to depth 2. Other structured read arrays default to 50 records, and document text fields default to 20,000 characters. JSON reports truncation. These limits cap output. REST still fetches complete collections when the server has no pagination.

For complete exports, specify the applicable unlimited flags:

```bash
craft list --limit 0
craft get page-123 --max-depth -1 --max-items 0 --max-chars 0
craft blocks get page-123 --depth -1 --max-items 0 --max-chars 0
craft collections items collection-123 --depth -1 --max-items 0 --max-chars 0
```

The `documents` resource group mirrors list, get, create, update, delete, move, clear and search. Existing top-level commands remain available. Each command's help contains a concrete example with synthetic IDs.

## Configuration and credential migration

`craft context` and `craft config inspect` report redacted effective settings, their sources, profile metadata and storage paths. API keys appear only as presence booleans. Connection URLs are redacted because link identifiers can grant access. Explicit REST URLs never inherit a saved profile's key.

REST URL precedence is `--api-url`, `CRAFT_API_URL`, then the selected or active REST profile. Key precedence is `--api-key`, `--api-key-env`, `CRAFT_API_KEY`, then the matching profile. MCP URL precedence is `--mcp-url`, `CRAFT_MCP_URL`, then the selected or active MCP profile.

Profile saves put URLs and API keys in a separate private credential generation. The metadata file points to its filename. Both files are mode 0600. The CLI writes a new credential generation, then switches the metadata reference atomically. If that switch fails, the old generation stays usable. Legacy inline profiles remain readable. Migrate existing profiles explicitly:

```bash
craft config migrate --dry-run
craft config migrate
craft context
```

Use profile commands to manage settings. Treat credential generations as private. An older CLI cannot read the new separated layout, so export any needed rollback configuration privately before migrating.

Noninteractive setup accepts a name and exactly one connection URL:

```bash
craft setup --name work --api-url "$CRAFT_API_URL" --api-key-env CRAFT_API_KEY
craft setup --name work-mcp --mcp-url "$CRAFT_MCP_URL"
```

## Local index

Sync is opt-in and stores private Craft content locally, under the configuration directory. It partitions snapshots by a hash of the exact REST URL and credential. A changed credential gets a separate snapshot. Sync refreshes documents whose modification time changed, re-fetches records without usable modification metadata, and removes documents no longer in the connection's listing. Successful reads are checkpointed; interrupted syncs must finish before local reads use the snapshot. `--full` refetches every document.

```bash
craft sync --profile work
craft list --profile work --data-source local --limit 10
craft get page-123 --profile work --data-source local
craft search 'release notes' --profile work --data-source local
craft cache status --profile work
craft cache clear --profile work --dry-run
craft cache clear --profile work --yes
```

The index uses private JSON snapshots and a persisted term index. Local search matches all query terms without contacting Craft. It supports document text, not regex, folder/date filters or block-scoped search. Unsupported filters fail explicitly. Only list, get and search accept local data. Native `craft local ...` commands still open the Craft application and are separate from this index.

`--data-source live` is the default and always contacts Craft. `local` explicitly permits an older snapshot. `auto` uses a complete snapshot at most five minutes old, then contacts Craft if it is absent, incomplete or older. Adjust that threshold with `--cache-max-age`. It never silently returns stale data after a live error. Remote mutations do not update the snapshot; rerun sync after edits. Raw upstream bytes require the live data source. Revoking a connection upstream does not erase an offline copy, so clear the cache when removing access.

## Embedded files

Structured JSON inputs expand files only with `--expand-files`. `@file://path` inserts parsed JSON, `@text://path` inserts UTF-8 text, and `@data://path` inserts base64. `@@` escapes one leading at sign. The total input and expanded JSON budget is 8 MiB. Only regular local files are accepted. Inserted file contents are data and do not trigger further expansion.

```bash
craft blocks add page-123 --json-file blocks.json --expand-files --dry-run
```

For example, a block's markdown value can be `"@text://notes.txt"`. Without the flag, that string remains literal. Use `--json-file` or stdin for payloads containing quotes or several blocks.

## Connection-specific API behavior

Use `craft search --date today 'notes'` for Daily Notes connections. Use `craft search --document-id page-123 'notes'` for Selected Documents connections. Existing `--document` selects block search on an All Documents connection. Tasks can use `--date` or `--location dailyNote`. Raw REST JSON remains available for connection-specific fields.

`craft folders update folder-123 --name Projects --dry-run` previews the native MCP folder metadata wrapper. Names, icons and tints use the verified MCP grammar. Experimental reminders depend on the connection's upstream availability. Task writes currently provide no revert payload; `tasks --save-revert` and `--diff` fail before writing. Block-level reversible edits remain available where Craft returns undo metadata.

File uploads default to the documented binary transport for general attachments. Images and videos retain detected media MIME types; `--content-type` remains available for an explicit choice. Live text uploads with automatic text/plain transport failed upstream, while the same connection accepted those bytes as application/octet-stream.
