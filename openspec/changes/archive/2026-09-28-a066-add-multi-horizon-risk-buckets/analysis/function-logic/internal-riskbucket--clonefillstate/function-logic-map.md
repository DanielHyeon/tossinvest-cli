# Function Logic Map: `cloneFillState`

- Source: `internal/riskbucket/fill.go`
- AST evidence: `ast.json` (5.7 pre-edit, extracted 2026-09-28 at HEAD `cf381447`)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit copy, `go test -coverpkg ./internal/journal,./internal/riskbucket` (journal untagged, riskbucket `tossos_testseams`); rows by `analysis/harness/preedit_rows.py`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in | a FillState | caller | pure copy |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.7 post-edit) |
|---|---|---|---|---|
| B1 | range at 444:2 | `for key, usage := range in.Buckets {`; then `usage.Latches = cloneLatchMap(usage.Latches)` (line last changed by `8b9821de`) | a066 commit | covered |
| B2 | range at 448:2 | `for id, order := range in.Orders {`; then `order.ReservedMinor = cloneMinorMap(order.ReservedMinor)` (line last changed by `8b9821de`) | a066 commit | covered |
| B3 | range at 452:3 | `for fillID, record := range order.Fills {`; then `record.TransferMinor = cloneMinorMap(record.TransferMinor)` (line last changed by `8b9821de`) | a066 commit | covered |
| B4 | range at 461:2 | `for latch, set := range in.OwnerLatches {`; then `out.OwnerLatches[latch] = set` (line last changed by `8b9821de`) | a066 commit | covered |
| B5 | if at 466:2 | `if in.SharedUsedMinor != nil {`; then `out.SharedUsedMinor = cloneMinorMap(in.SharedUsedMinor)` (line last changed by `00000000`) | a066: 5.7: shared usage map copied when present; nil stays nil (crash-pure DeepEqual contract) | covered |

5.7 post-edit (HEAD `54e67495` + 5.7 working tree): 4 → 5: new B5 copies `SharedUsedMinor` when non-nil (nil stays nil — crash-pure DeepEqual contract). Pre-edit table: `analysis/pre-edit/5.7/internal-riskbucket--clonefillstate.md`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| cloneLatchMap / cloneMinorMap / cloneActualFillEvidence | deep copy | none | AST |

## State mutations and fallbacks

- No broker call; the caller's transaction owns every write.

## Safety conclusion

- Safe edit boundary: copy the new `SharedUsedMinor` map so `unchanged` and `next` do not share it.
- High-risk impact: yes — fill accounting (체결 반영). The fill and Position are never rejected by this change.
