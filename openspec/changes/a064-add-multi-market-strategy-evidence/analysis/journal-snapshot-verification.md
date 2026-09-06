# Journal Snapshot-Only Lineage Verification

- Date: 2026-08-04
- Scope: trading-journal v21 lineage and dormant immutable snapshot reads only
- Safety boundary: no Guardian, dispatch, broker, apply-hook, LIVE or operating-toggle integration

## RED

Before production code was added, the focused command failed to compile because `schemaV21`, the consumed
snapshot lineage fields, `NewStrategyEvidenceReadBoundary`, `NewDormantSnapshotReadPort`,
`SnapshotReference` and `ErrSnapshotUnavailable` did not exist:

```text
go test -count=1 ./internal/journal ./internal/strategyevidence
FAIL (undefined-contract compile errors)
```

## GREEN

```text
go test -run 'TestStrategyEvidence|TestMigrationV21|TestDormantSnapshot|TestOpenReadOnlyRejectsDamagedV21' -count=1 ./internal/journal ./internal/strategyevidence
ok internal/journal 7.068s
ok internal/strategyevidence 0.993s

go test -timeout=5m -count=1 ./internal/journal ./internal/strategyevidence
ok internal/journal 214.081s
ok internal/strategyevidence 0.916s

go test -race -timeout=5m -run 'TestStrategyEvidence|TestMigrationV21|TestDormantSnapshot|TestOpenReadOnlyRejectsDamagedV21' -count=1 ./internal/journal ./internal/strategyevidence
ok internal/journal 27.317s
ok internal/strategyevidence 2.065s

go vet ./internal/journal ./internal/strategyevidence
PASS

openspec validate a064-add-multi-market-strategy-evidence --strict --no-interactive
Change 'a064-add-multi-market-strategy-evidence' is valid
```

The full two-package race command was also attempted with a ten-minute timeout. `internal/strategyevidence`
passed under race (`1.404s`), while `internal/journal` timed out after ten minutes during its large SQLite
migration-heavy suite. The timeout stack showed active migration work and parallel tests waiting for the
suite barrier, not a deadlock or race report. Targeted race coverage for every a064 journal/evidence test
passes as recorded above; the full non-race package suite passes.

## Isolation and no mutation

- `schemaV21` contains two `ALTER TABLE ... ADD COLUMN ... TEXT` statements **and two `CREATE TRIGGER`**
  statements (`internal/journal/strategy_evidence.go:14-33`). An earlier revision of this line said "only
  two ALTER TABLE", which made the isolation argument rest on a description the migration does not match;
  review.md described the triggers, so the two documents disagreed. Corrected 2026-09-07 (issues.md I23).
- Migration/schema tests reject evidence payload, revision, credential, secret, source-response and evidence
  table storage in the trading journal. Scope limit measured 2026-09-07: that test greps the `schemaV21`
  **string literal in its own package**, not `sqlite_master` of the migrated database, so it cannot see a
  payload or credential table introduced by any other migration (issues.md I14).
- Dormant read AST tests reject `Exec`, `ExecContext`, `Begin`, `BeginTx`, mutating SQL, `net/http`, broker,
  dispatch, execution-gateway, Guardian, runtime, operating and toggle imports. Scope limit measured
  2026-09-07: each test parses one file and walks one method body, and matches SQL literals with a trailing
  space, so a write reached through a same-package helper, a differently named method, a `driver.Conn` or
  `DELETE\nFROM` passes cleanly (issues.md I12).
- After a journal snapshot-lineage read, `intents`, `mutation_attempts` and `risk_reservations` remain zero.
- Exact KR and US replay is market-qualified; failure in one market does not alter or gate the other.

## Independent HIGH integrity RED to GREEN

The review RED removed only the append-only trigger in a temporary evidence.db, changed otherwise-valid
Header data behind an already sealed snapshot, and replayed the original ID/digest. Before the fix, all six
cases returned success: symbol, issuer, mapping, cross-market scope, future market-effective date and future
source/availability/observed/ingested timestamps. Separate journal REDs showed a direct partial ID-only row
was accepted and an unsupported market was returned.

GREEN changes and guarantees:

- `snapshotDigest` length-prefixes the normalized query and every immutable Header field — EvidenceID,
  market/symbol/issuer/mapping, kind/schema/authority/source and revision identities, effective date, four
  timestamps, currency/unit/availability/confidence — followed by payload digest.
- `DormantSnapshotReadPort.Replay` independently validates exact scope, both as-of cutoffs, source event and
  market-local effective date before accepting an item, then recomputes the full-Header digest.
- v21 INSERT/UPDATE triggers accept either two NULLs or an exact lowercase 64-hex ID/digest pair. The UPDATE
  RED drops the older blanket immutability trigger first, proving the new pair guard is independently active.
- Journal reads refuse any market other than exact lineage values `KR` and `US`.

The focused non-race and race commands above include these HIGH regression tests and pass.

## Remaining repository gate

Function Logic Maps for the three a064-modified existing functions were refreshed against current AST and their
branch maps record GREEN coverage.

Superseded 2026-09-07 by the completion pass. The comparison base was moved from `c57915dd` to `bc03c4d4`
because 307 unrelated commits had made the checker attribute 341 foreign functions to a064; at the new base
it reports 0 required functions, so the checker's silence here is vacuity, not coverage (issues.md I27).
Repository-wide `make sdd-check`, `make test`, `make test-seams`, `make test-race`, `make vet` and
`make validate` all pass as of that date. What blocks completion is not the gates but the recorded claims:
the branch maps' GREEN coverage above is contradicted for rows B1-B3 of `snapshotDigest`, B1-B2 of
`insertExactStrategyDecision`, and B1/B6/B11/B13/B17/B21/B23/B24/B27 of `ReadOnly.checkSchema` — see
issues.md I1-I9. `make gate CHANGE=a064-add-multi-market-strategy-evidence` is not run.
