# Agent CLI behavior audit

Use the local cli-best-practices Audit v3 workflow and checklist. Score each numbered check against the built binary using fixtures or local dry-runs. Do not treat `craft audit agent-dx` or flag presence as evidence.

The v3 source at 24b9960 labels itself 85 checks but lists 97. Report the real denominator, pass/fail evidence and declared exemptions. Read the 2026-09-26 baseline and final reports in docs/plans. Do not copy a hard-coded perfect score.

Run `go test ./...`, `go vet ./...`, and `python3 scripts/verify_agent_contract.py /tmp/craft-cli` after building. For CLI Spec claims, also record the external checker result and the safe command used. Never mutate production data to score a check.
