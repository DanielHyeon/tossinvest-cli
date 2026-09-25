# Function Logic Map: `ApplyFill`

- Source: `internal/riskbucket/fill.go`
- AST evidence: `ast.json` (extracted Wave 2A 2026-09-25 at HEAD `648df8ef`; 96–273, 38 branches, 23 returns;
  byte-identical to the archived a072 bundle `internal-riskbucket--applyfill/ast.json`)
- Risk scan: `risk-pattern-report.md`

## Why this bundle is new in Wave 2A

`ApplyFill` was written by a066 Wave 1A (`8b9821de`, before base `23794f86`) and changed by a066 Wave 1D
(`4a364caf`: order_key identity and per-decision target HELD). Because it existed at base, the Wave 1D edit is a
modified existing function and needed a bundle; none was written. HEAD's body equals `4a364caf`'s.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | bucket usages, per-order watermarks, owner latches | journal reconstruction | input never mutated (deep copy) |
| `event` identity/quantity | non-empty fill/order id, 0 < cumulative ≤ ordered, policy digest, canonical currencies | journal fill sidecar | typed `FILL_EVIDENCE_INCONSISTENT` refusal, unchanged copy |
| reserved/target HELD minor | parseable non-negative integers per bucket; target map same size as reserved | registered risk order rows | refusal, unchanged copy |
| actual evidence | optional; price/fee/FX pair must match order currencies | persisted actual evidence | unknown → `UNKNOWN_ACTUAL_RISK` latch, fill still applied |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage at `648df8ef` |
|---|---|---|---|---|
| B1 | if at 100:2 | `if event.FillID == "" \|\| event.OrderID == "" \|\| event.OrderQuantity == 0 \|\| event.NewCumulativeFill == 0 \|\| event.NewCumulativeFill > event.OrderQuantity \|\| event.ReservationPolicyDigest == "" \|\|`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "fill_identity_or_quantity", nil)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B2 | if at 104:2 | `if err := validateFillBuckets(next.Buckets, event.ReservedMinor); err != nil {`; then `return unchanged, result, err` (line last changed by `8b9821de`) | a066 commit | covered |
| B3 | if at 107:2 | `if len(event.TargetHeldMinor) != 0 {`; then `if len(event.TargetHeldMinor) != len(event.ReservedMinor) {` (line last changed by `4a364caf`) | a066: owner-wide multi-decision fill carries per-decision target HELD (Wave 1D) | covered |
| B4 | if at 108:3 | `if len(event.TargetHeldMinor) != len(event.ReservedMinor) {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "target_held_bucket_count", nil)` (line last changed by `4a364caf`) | a066: target HELD bucket count differs from reserved buckets: refuse, state unchanged | NOT covered |
| B5 | range at 111:3 | `for key := range event.ReservedMinor {`; then `if _, err := parseMinor(event.TargetHeldMinor[key], 0); err != nil {` (line last changed by `4a364caf`) | a066: iterate reserved buckets to validate every target HELD value | covered |
| B6 | if at 112:4 | `if _, err := parseMinor(event.TargetHeldMinor[key], 0); err != nil {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "target_held_usage", err)` (line last changed by `4a364caf`) | a066: unparseable target HELD: refuse, state unchanged | NOT covered |
| B7 | if at 118:2 | `if strings.TrimSpace(event.OrderKey) != "" {`; then `orderIdentity = event.OrderKey` (line last changed by `4a364caf`) | a066: immutable order_key (not the broker order id) keys the order watermark (Wave 1D) | covered |
| B8 | if at 122:2 | `if !exists {`; then `order = OrderFillState{` (line last changed by `8b9821de`) | a066 commit | covered |
| B9 | else at 135:9 | `} else if order.OrderQuantity != event.OrderQuantity \|\| order.QuoteCurrency != event.QuoteCurrency \|\| order.BaseCurrency != event.BaseCurrency \|\| order.ReservationPolicyDigest != event.ReservationPolicyDigest \|\| !equalMinorMaps(order.ReservedMinor, event.ReservedMinor) {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "order_reservation", nil)` (line last changed by `8b9821de`) | a066 commit | covered |
| B10 | range at 132:3 | `for key := range event.ReservedMinor {`; then `order.TransferredMinor[key] = "0"` (line last changed by `8b9821de`) | a066 commit | covered |
| B11 | if at 135:9 | `} else if order.OrderQuantity != event.OrderQuantity \|\| order.QuoteCurrency != event.QuoteCurrency \|\| order.BaseCurrency != event.BaseCurrency \|\| order.ReservationPolicyDigest != event.ReservationPolicyDigest \|\| !equalMinorMaps(order.ReservedMinor, event.ReservedMinor) {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "order_reservation", nil)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B12 | if at 139:2 | `if record, seen := order.Fills[event.FillID]; seen {`; then `if record.CumulativeFill != event.NewCumulativeFill {` (line last changed by `8b9821de`) | a066 commit | covered |
| B13 | if at 140:3 | `if record.CumulativeFill != event.NewCumulativeFill {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "fill_identity_watermark", nil)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B14 | if at 143:3 | `if record.ActualKnown {`; then `result.Duplicate = true` (line last changed by `8b9821de`) | a066 commit | covered |
| B15 | if at 148:3 | `if !known {`; then `result.Duplicate = true` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B16 | range at 152:3 | `for key, transferRaw := range record.TransferMinor {`; then `transfer, err := parseMinor(transferRaw, 0)` (line last changed by `8b9821de`) | a066 commit | covered |
| B17 | if at 154:4 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "record_transfer", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B18 | if at 158:4 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "record_filled", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B19 | if at 162:4 | `if actualMinor.Cmp(target) > 0 {`; then `target.Set(actualMinor)` (line last changed by `8b9821de`) | a066 commit | covered |
| B20 | if at 165:4 | `if target.Cmp(previousFilled) > 0 {`; then `delta := new(big.Int).Sub(target, previousFilled)` (line last changed by `8b9821de`) | a066 commit | covered |
| B21 | if at 169:5 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B22 | if at 173:5 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage_overflow", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B23 | if at 185:3 | `if err := recomputeOverageLatches(&next); err != nil {`; then `return unchanged, FillResult{}, err` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B24 | if at 193:2 | `if event.NewCumulativeFill <= order.CumulativeFill {`; then `return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "cumulative_fill", nil)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B25 | if at 206:2 | `if actualKnown {`; then `record.ActualEvidence = cloneActualFillEvidence(event.Actual)` (line last changed by `8b9821de`) | a066 commit | covered |
| B26 | range at 209:2 | `for key, reservedRaw := range event.ReservedMinor {`; then `reserved, err := parseMinor(reservedRaw, 0)` (line last changed by `8b9821de`) | a066 commit | covered |
| B27 | if at 211:3 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "reserved_minor", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B28 | if at 216:3 | `if err != nil \|\| allocated.Cmp(previousTransferred) < 0 {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "transferred_minor", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B29 | if at 222:3 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "held_usage", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B30 | if at 226:3 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage", err)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B31 | if at 230:3 | `if len(event.TargetHeldMinor) != 0 {`; then `targetHeld, targetErr := parseMinor(event.TargetHeldMinor[key], 0)` (line last changed by `4a364caf`) | a066: target HELD present for this bucket | covered |
| B32 | if at 232:4 | `if targetErr != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "target_held_usage", targetErr)` (line last changed by `4a364caf`) | a066: unparseable target HELD at transfer time: refuse, state unchanged | NOT covered |
| B33 | if at 235:4 | `if heldDeduction.Cmp(targetHeld) > 0 {`; then `heldDeduction.Set(targetHeld)` (line last changed by `4a364caf`) | a066: transfer exceeds the decision's own HELD: cap the deduction and latch RISK_OVERAGE instead of consuming another decision's HELD | covered |
| B34 | if at 241:3 | `if heldDeduction.Cmp(held) > 0 {`; then `heldDeduction.Set(held)` (line last changed by `8b9821de`) | a066 commit | NOT covered |
| B35 | if at 248:3 | `if actualKnown && actualMinor.Cmp(filledDelta) > 0 {`; then `filledDelta.Set(actualMinor)` (line last changed by `8b9821de`) | a066 commit | covered |
| B36 | if at 252:3 | `if err != nil {`; then `return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage_overflow", err)` (line last changed by `8b9821de`) | a066 commit | covered |
| B37 | if at 256:3 | `if !actualKnown {`; then `latchUsage(&usage, LatchUnknownActualRisk)` (line last changed by `8b9821de`) | a066 commit | covered |
| B38 | if at 268:2 | `if err := recomputeOverageLatches(&next); err != nil {`; then `return unchanged, FillResult{}, err` (line last changed by `8b9821de`) | a066 commit | NOT covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateFillBuckets`, `parseMinor` | bucket set and minor-unit parsing | any error returns the unchanged copy | source |
| `actualFillMinor` | actual monetary exposure from price/fee/FX | unknown is a latch, never 0 | source |
| `proportionalAllocation`, `addMinor` | proportional HELD transfer and bounded 256-bit addition | overflow is a refusal | source; `TestApplyFillRejectsStoredMinorOverflowWithoutPartialMutation` |
| `latchUsage`, `recomputeOverageLatches`, `clearResolvedUnknownLatches` | bucket/owner latches | recompute error returns unchanged copy | source |

Production callers (CodeGraph 1.6.0): `applyRiskBucketFillInTx` and `completeRiskBucketFillActual` in
`internal/journal/risk_bucket_fill.go`; 13 test callers.

## State mutations and fallbacks

- Pure: returns a deep-copied next state; every error path returns an unchanged deep copy so the journal
  commits all bucket changes or none.
- Duplicate fill id with the same cumulative quantity is idempotent; late actual evidence completes
  `filled = max(transfer, actual)` monotonically.

## Safety conclusion

- High-risk impact: yes — monetary exposure accounting that gates new entries.
- Conservative direction only: shortfall of the decision's own HELD caps the deduction and latches
  `RISK_OVERAGE` (B33) instead of borrowing another decision's HELD.
- Measured residual risk: 19 of 38 branch bodies are not executed by any test, almost all defensive
  corrupt-state parse exits, plus B1 (invalid identity/quantity refusal) and B38 (latch recompute error).
