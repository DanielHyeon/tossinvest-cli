# Function Logic Map: `recomputeOverageLatches`

- Source: `internal/riskbucket/fill.go`
- AST evidence: `ast.json` (5.7 pre-edit, extracted 2026-09-28 at HEAD `cf381447`)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit copy, `go test -coverpkg ./internal/journal,./internal/riskbucket` (journal untagged, riskbucket `tossos_testseams`); rows by `analysis/harness/preedit_rows.py`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| state.Buckets | owner-only usage per bucket key with the snapshot limit | journal loader | parse/overflow → FILL_EVIDENCE_INCONSISTENT refusal (caller latches, fill kept) |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.7 post-edit) |
|---|---|---|---|---|
| B1 | range at 356:2 | `for key, usage := range state.Buckets {`; then `limit, err := parseMinor(usage.LimitMinor, 0)` (line last changed by `8b9821de`) | a066 commit | covered |
| B2 | if at 358:3 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_limit", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B3 | if at 362:3 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_held", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B4 | if at 366:3 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_filled", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B5 | if at 370:3 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_usage_overflow", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B6 | if at 375:3 | `if shared, ok := state.SharedUsedMinor[key]; ok {`; then `others, err := parseMinor(shared, 0)` (line last changed by `00000000`) | a066: 5.7: other entries' ledger usage in a shared bucket is added before comparing with the limit (design D5 bucket sum) | covered |
| B7 | if at 377:4 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_shared_usage", err)` (line last changed by `00000000`) | not a066 | NOT covered |
| B8 | if at 380:4 | `if used, err = addMinor(used, others, 0); err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_shared_usage_overflow", err)` (line last changed by `00000000`) | not a066 | NOT covered |
| B9 | if at 385:3 | `if overage.Sign() > 0 {`; then `anyOverage = true` (line last changed by `8b9821de`) | a066 commit | covered |
| B10 | if at 388:4 | `if err != nil {`; then `return refusal(RefusalFillEvidenceInconsistent, "overage_previous", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B11 | if at 391:4 | `if overage.Cmp(previous) > 0 {`; then `usage.OverageMinor = overage.String()` (line last changed by `8b9821de`) | a066 commit | covered |
| B12 | if at 397:2 | `if !anyOverage {`; then `return nil` (line last changed by `8b9821de`) | a066 commit | covered |
| B13 | range at 401:2 | `for key, usage := range state.Buckets {`; then `latchUsage(&usage, LatchRiskOverage)` (line last changed by `8b9821de`) | a066 commit | covered |

5.7 post-edit (HEAD `54e67495` + 5.7 working tree): 10 → 13: new B6–B8 add `SharedUsedMinor[key]` (other entries' ledger usage) to `used` before the limit comparison. Pre-edit table: `analysis/pre-edit/5.7/internal-riskbucket--recomputeoveragelatches.md`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parseMinor / addMinor | exact minor arithmetic | refusal on invalid | AST |
| latchUsage | latch every owner bucket when any bucket is over | pure | AST |

## State mutations and fallbacks

- No broker call; the caller's transaction owns every write.

## Safety conclusion

- Safe edit boundary (5.7 shared overage): add this transaction's `SharedUsedMinor[key]` (other entries' ledger usage) to `used` before comparing with the limit (design D5: the bucket sum). The owner-only comparison is the measured defect (shared usage 128 > 100, zero RISK_OVERAGE).
- High-risk impact: yes — fill accounting (체결 반영). The fill and Position are never rejected by this change.
