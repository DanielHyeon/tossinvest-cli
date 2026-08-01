# a049 implementation handoff

## Implemented in this worktree

- Isolated, rebuildable `performance.db` schema v1 with phase-SIGKILL-proven atomic migration, private files, exact compare-and-append trade/observation/measurement writes, 90-day retention, 24-hour cadence after backlog drain, and bounded 500-row prune transactions (four transactions/run before immediate reschedule).
- Exact typed lineage model with distinct `complete` and `link_missing` states. No symbol/time approximation API exists.
- Existing-observation-only BUY/SELL-consistent 5/15/30 minute gross/cost markout, slippage, MFE/MAE, cost-adjusted lane/policy aggregates, exact observation source/time/version provenance, metric definitions/units/sample/period/source/version, and explicit `not_measured`/`insufficient_sample` states.
- Actual Dashboard SQL is bounded to 90 days and 10,000 trades, resolves latest snapshots only inside the filtered CTE, applies identical state-count filters, and is EXPLAIN/p95 tested with 1,000,000 raw rows.
- Session-gated `/performance-history` UI with fixed server query defaults and no form, input, button, arbitrary text/number entry, mutation seam, order capability, lane toggle, or LIVE approval capability.
- Legacy portfolio aggregate JSON byte regression and dependency-closure checks.

## Deliberately deferred integration

Task 2.1 remains unchecked. a045 owns journal schema v13 and a047 may add strategy provenance, so this branch does not reserve v14/v15, alter `journal.SchemaVersion`, change `RiskIntent`/`Decision` JSON bytes, or guess a cross-version join. After a045 and a047 are integrated, allocate the then-current next journal version and implement `JournalLineageReader.ClosedStrategyTrades` using persisted identifiers only. That adapter must complete the nullable journal lineage migration before a049's independent review/gate can close.

## Verification evidence

- `go test ./internal/performance ./internal/console ./internal/journal -count=1` — pass.
- `go test ./... -count=1` — pass.
- `go vet ./...` — pass.
- `go test -race ./internal/performance -count=1` — pass; the 1M wall-clock benchmark is skipped only under race instrumentation and remains enforced by the normal suite.
- `go test -race ./internal/console -count=1` — pass.
- `go test -race ./internal/journal -run '^TestA049DoesNotChangePortfolioAggregateBytes$' -count=1` — pass. The repository-wide journal race suite was not retried because its modernc SQLite parallel fixture is a known timeout; an earlier combined run was manually interrupted and is not treated as a product failure.
- `GOOS=windows GOARCH=amd64 go test -exec=/bin/true ./internal/performance ./internal/console ./internal/journal` — pass.
- `openspec validate a049-add-lane-performance --strict --no-interactive` — pass.

`make sdd-sync`, `make sdd-check`, `make gate`, independent review, commit, integration, push, and deployment are intentionally not performed in this worktree before the coordinating agent's approval. The persisted `base-commit.txt` still predates the branch base, so Function Logic Map checking will remain fail-closed until the coordinator performs the approved rebase/base-marker workflow.
