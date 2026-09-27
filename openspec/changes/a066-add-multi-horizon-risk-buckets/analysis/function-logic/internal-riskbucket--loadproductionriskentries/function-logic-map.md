# Function Logic Map: `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- AST evidence: `ast.json` (5.6.1 pre-edit, extracted 2026-09-27 at HEAD `f2decd0a` before any 5.6.1 edit)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit source copy (`analysis/harness/test_in_copy.sh`), `go test -coverpkg ./internal/journal,./internal/riskbucket` over the journal package (untagged) and riskbucket (untagged + `tossos_testseams`); rows rendered by `analysis/harness/preedit_rows.py` from `git show f2decd0a:<file>`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| config, scope, reserve, limits | sealed production authority; journal at schema `productionRiskJournalSchema` (=27) | read-only journal | any failure → no snapshot (entry closed) |

## Branches and early returns

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 340:2 | `if !ok {` | NOT covered |
| B2 | if at 343:2 | `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | NOT covered |
| B3 | if at 352:2 | `if err != nil {` | NOT covered |
| B4 | if at 357:2 | `if err := db.PingContext(ctx); err != nil {` | NOT covered |
| B5 | if at 361:2 | `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \|\| version != productionRiskJournalSchema {` | NOT covered |
| B6 | if at 365:2 | `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` | NOT covered |
| B7 | if at 376:2 | `if authorityObserved.After(scope.AsOf) \|\| authorityFresh.Before(scope.AsOf) {` | NOT covered |
| B8 | range at 380:2 | `for _, dimension := range requiredDimensions {` | covered |
| B9 | if at 382:3 | `if err != nil {` | NOT covered |
| B10 | if at 386:3 | `if err != nil {` | covered |
| B11 | if at 395:3 | `if err != nil {` | NOT covered |
| B12 | if at 405:3 | `if err != nil {` | NOT covered |
| B13 | if at 414:3 | `if err != nil {` | NOT covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readProductionRiskUsage` + `aggregateProductionRiskUsage` | per-dimension usage | error / latched → refuse | AST B9, B10 |

## State mutations and fallbacks

- Read-only journal handle; builds sealed snapshot entries.

## Safety conclusion

- Safe edit boundary (5.6.1 F1): call `ReadJournalBucketUsage` (the shared function) and refuse when `Latched` with the unchanged message. Note: B5 pins the journal schema to exactly 27, so on any current journal (v32–v34) production snapshot loading already fails closed — pre-existing, reported, not changed here.
- High-risk impact: yes — q_final sizing/admission authority.
