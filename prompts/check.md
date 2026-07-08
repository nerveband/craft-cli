# Quick Check Prompt

Copy-paste this into Claude Code to verify everything builds, tests pass, and docs are up to date:

---

Run the full local verification pass and docs/content audit. Do these steps:

**Build checks:**
1. Run `go test ./...` for all tests
2. Run `go vet ./...` for static analysis
3. Run `go build -o craft-cli .` to verify the binary builds
4. Run `./craft-cli version` to confirm version output
5. Run `./craft-cli info` to verify API connectivity

**API integration checks:**
6. Test core read operations against the live API:
   - `./craft-cli list` — verify document listing works
   - `./craft-cli folders` — verify folder listing works
   - `./craft-cli search "test"` — verify search works
7. If staging credentials or test workspace is available, run broader integration tests:
   - `./craft-cli get <known-doc-id>` — verify document retrieval
   - `./craft-cli blocks get <known-doc-id>` — verify block-level access
   - `./craft-cli collections` — verify collections listing

**Docs & content audit:**
8. Check `git diff --name-only HEAD~5` to see what changed recently
9. For any CLI command changes: check if `README.md` and `docs/llm/` need updating
10. For any API client changes: check if `internal/api/client_test.go` covers the new behavior
11. For any output format changes: verify all format modes still work (`--format json`, `--format table`, `--format markdown`)

**Agent DX audit:**
12. Run the built-in scorecard:
    - `./craft-cli audit agent-dx --format json`
    - Confirm the score is 85/85 on the v2 audit and there is no regression from the previous run
13. Run live checks only when explicitly configured:
    - `CRAFT_LIVE_TESTS=1 CRAFT_LIVE_REST_URL=... CRAFT_LIVE_MCP_URL=... go test ./... -run Live`
    - `CRAFT_LIVE_TESTS=1 CRAFT_LIVE_WRITEONLY_URL=... go test ./internal/api -run LiveRESTWriteOnlyState`
    - `CRAFT_LIVE_TESTS=1 CRAFT_LIVE_MUTATION_TESTS=1 CRAFT_LIVE_REST_URL=... go test ./internal/api -run LiveRESTMutation`
14. Check the change against the repo's agent-DX goals:
    - All commands produce structured JSON by default
    - Errors are structured and actionable (not just prose)
    - `--dry-run` exists for destructive commands
    - `--quiet`, `--json-errors`, `--output-only`, `--id-only` flags work correctly
    - Raw payload input via stdin is supported where applicable
    - Progressive help discovery works (`craft`, `craft <cmd>`, `craft <cmd> --help`)

**Reporting:**
15. Report what passed, what failed, what was skipped, and any agent-DX regressions or wins.
