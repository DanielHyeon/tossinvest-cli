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

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.6.1 post-edit) |
|---|---|---|---|---|
| B1 | if at 467:2 | `if err != nil {`; then `return riskBucketOrderAuthority{}, err` (line last changed by `c60fee07`) | a066: 5.6.1 F2 query: v34 rows join their own record (UNION ALL legacy key join for NULL rows); query error returns | NOT covered |
| B2 | range at 473:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {`; then `required[dimension] = true` (line last changed by `c60fee07`) | a066 commit | covered |
| B3 | for at 477:2 | `for rows.Next() {`; then `var binding riskBucketOrderAuthorityBinding` (line last changed by `c60fee07`) | a066 commit | covered |
| B4 | if at 480:3 | `if err := rows.Scan(&binding.Dimension, &binding.Value, &binding.PolicyVersion, &binding.ReservationID, &binding.ReservedMinor, &binding.SnapshotID, &binding.SnapshotDigest, &binding.PolicyDigest, &policyDigest, &binding.PolicyRecordDigest, &rowQuote, &rowBase); err != nil {`; then `return riskBucketOrderAuthority{}, fmt.Errorf("%w: order authority scan: %v", ErrRiskBucketReplayMismatch, err)` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B5 | if at 484:3 | `if _, duplicate := byDimension[dimension]; !required[dimension] \|\| duplicate \|\| binding.PolicyDigest != policyDigest \|\| rowQuote == "" \|\| rowBase == "" \|\| (quote != "" && quote != rowQuote) \|\| (base != "" && base != rowBase) {`; then `return riskBucketOrderAuthority{}, fmt.Errorf("%w: non-canonical order authority", ErrRiskBucketSnapshotMismatch)` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B6 | if at 490:2 | `if err := rows.Err(); err != nil {`; then `return riskBucketOrderAuthority{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B7 | if at 493:2 | `if len(byDimension) != len(required) {`; then `return riskBucketOrderAuthority{}, fmt.Errorf("%w: order authority dimension set", ErrRiskBucketSnapshotMismatch)` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B8 | range at 497:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {`; then `binding, ok := byDimension[dimension]` (line last changed by `c60fee07`) | a066 commit | covered |
| B9 | if at 499:3 | `if !ok {`; then `return riskBucketOrderAuthority{}, fmt.Errorf("%w: missing %s order authority", ErrRiskBucketSnapshotMismatch, dimension)` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B10 | if at 505:2 | `if err != nil {`; then `return riskBucketOrderAuthority{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |

5.6.1 post-edit (HEAD `b8211926`): 10 → 10 branches; only the query changed (own record for v34 rows, legacy key join for NULL rows). Pre-edit table: `analysis/pre-edit/5.6.1/internal-journal--loadriskbucketorderauthority.md`.

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
