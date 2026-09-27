# Function Logic Map: `readProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- AST evidence: `ast.json` (5.6.1 pre-edit, extracted 2026-09-27 at HEAD `f2decd0a` before any 5.6.1 edit)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit source copy (`analysis/harness/test_in_copy.sh`), `go test -coverpkg ./internal/journal,./internal/riskbucket` over the journal package (untagged) and riskbucket (untagged + `tossos_testseams`); rows rendered by `analysis/harness/preedit_rows.py` from `git show f2decd0a:<file>`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| db, account, dimension, value | read-only journal | risk_bucket_reservations (+snapshot, policy joins) | query/scan error returned |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.7 post-edit) |
|---|---|---|---|---|
| B1 | if at 457:2 | `if err != nil {`; then `return nil, err` (line last changed by `8022f578`) | not a066 | NOT covered |
| B2 | for at 462:2 | `for rows.Next() {`; then `var row productionRiskUsageRow` (line last changed by `8022f578`) | not a066 | covered |
| B3 | if at 464:3 | `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,`; then `return nil, err` (line last changed by `8022f578`) | not a066 | NOT covered |

5.7 post-edit (HEAD `54e67495` + 5.7 working tree): 3 → 3; body unchanged (file lines shifted).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryContext` | reservation rows of one bucket | error returned | AST |

## State mutations and fallbacks

- Read-only.

## Safety conclusion

- Safe edit boundary (5.6.1 F1): parameter type `*sql.DB` → `UsageQueryer` (satisfied by *sql.DB and *sql.Tx) so the admission transaction can call the same reader. Body unchanged.
- High-risk impact: yes — q_final sizing/admission authority.
