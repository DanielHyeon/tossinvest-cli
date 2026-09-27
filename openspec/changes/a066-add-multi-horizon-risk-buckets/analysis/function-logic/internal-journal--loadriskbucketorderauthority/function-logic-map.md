# Function Logic Map: `loadRiskBucketOrderAuthority`

- Source: `internal/journal/risk_bucket_fill.go`
- AST evidence: `ast.json` (5.6.1 pre-edit, extracted 2026-09-27 at HEAD `f2decd0a` before any 5.6.1 edit)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit source copy (`analysis/harness/test_in_copy.sh`), `go test -coverpkg ./internal/journal,./internal/riskbucket` over the journal package (untagged) and riskbucket (untagged + `tossos_testseams`); rows rendered by `analysis/harness/preedit_rows.py` from `git show f2decd0a:<file>`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| decisionID | an admitted q_final decision | risk_bucket_reservations/snapshots/policies | scan error → ErrRiskBucketReplayMismatch; shape error → ErrRiskBucketSnapshotMismatch |

## Branches and early returns

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 462:2 | `if err != nil {` | NOT covered |
| B2 | range at 468:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | covered |
| B3 | for at 472:2 | `for rows.Next() {` | covered |
| B4 | if at 475:3 | `if err := rows.Scan(&binding.Dimension, &binding.Value, &binding.PolicyVersion, &binding.ReservationID, &binding.ReservedMinor, &binding.SnapshotID, &binding.SnapshotDige` | NOT covered |
| B5 | if at 479:3 | `if _, duplicate := byDimension[dimension]; !required[dimension] \|\| duplicate \|\| binding.PolicyDigest != policyDigest \|\| rowQuote == "" \|\| rowBase == "" \|\| (quot` | NOT covered |
| B6 | if at 485:2 | `if err := rows.Err(); err != nil {` | NOT covered |
| B7 | if at 488:2 | `if len(byDimension) != len(required) {` | NOT covered |
| B8 | range at 492:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | covered |
| B9 | if at 494:3 | `if !ok {` | NOT covered |
| B10 | if at 500:2 | `if err != nil {` | NOT covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQL join reservations→snapshots→policies **by key only** | read the five bindings, currencies and policy digest | a second record under the same key would join the first entry's record | AST (query at the head) |
| `riskBucketRecordDigest` | order authority digest | error returns | AST |

## State mutations and fallbacks

- Read-only inside the caller transaction.

## Safety conclusion

- Safe edit boundary (5.6.1 F2): v34 rows join `risk_bucket_policy_records` by `(key, r.policy_record_digest)`; pre-v34 rows (NULL) keep the key join. Branches unchanged.
- High-risk impact: yes — q_final sizing/admission authority.
