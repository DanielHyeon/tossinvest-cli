# Function Logic Map: `ApplyFill`

- Source: `internal/riskbucket/fill.go`
- Qualified function: `ApplyFill`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| prior fill state | exact five dimensions, nonnegative canonical minor strings and immutable order/fill watermarks | durable journal projection | unchanged deep copy plus typed refusal |
| fill identity | nonempty fill/order, positive monotonic cumulative quantity no greater than order quantity | broker fill plus durable scoped order binding | unchanged state |
| reservation binding | exact policy digest, currencies and per-bucket reserved/optional target-held maps | original q_final reservation | unchanged state |
| actual risk evidence | optional exact price/fee/FX evidence | official fill/cost/FX provenance | missing evidence is preserved and latches future entry, not the fill |

## Branches and early returns

| Branch family | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | invalid identity, currency, bucket shape or target-held map | none outside cloned next state | evidence-inconsistent refusal | validation matrix |
| B7-B11 | choose immutable order key; create first order state or reject divergent replay | initialize zero transferred minors in clone | success/refusal | order-key isolation tests |
| B12-B22 | duplicate fill ID | exact duplicate no-op, or late actual-evidence completion that only increases filled risk and recomputes latches | unchanged/error/completed result | idempotency and late-evidence tests |
| B23 | cumulative watermark does not advance | none | evidence-inconsistent refusal | nonmonotonic replay test |
| B24-B37 | allocate delta across every bucket | reduce HELD conservatively, increase FILLED, cap deduction, latch overage/unknown actual; any parse/overflow error returns original clone | success/refusal | partial fill, overage, unknown actual and overflow suites |
| B38 | final overage-latch recomputation fails | discard cloned mutations | error | injected invalid-state rollback test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cloneFillState` | guarantee all-or-none pure transition | both returned states are deep copies | immutability tests |
| `validateFillBuckets` / `parseMinor` / `addMinor` | exact dimensional and arbitrary-precision integer checks | no float, truncation or repair | fill validation/property tests |
| `proportionalAllocation` | convert cumulative quantity into monotonic reservation transfer | deterministic integer allocation | partial-fill tests |
| `actualFillMinor` | bind optional actual price/fees/FX | unknown remains explicit | actual-risk tests |
| `recomputeOverageLatches` / `clearResolvedUnknownLatches` | preserve future-entry latches from complete durable evidence | never discard fill | latch tests |

## State mutations and fallbacks

- Operates on a deep-cloned in-memory state; callers atomically persist the returned next state or none.
- A fill is never rejected merely because actual risk is unknown: the fill remains authoritative while
  `UNKNOWN_ACTUAL_RISK` blocks new exposure. Overage similarly latches without deleting or downsizing the fill.
- No market substitution, currency conversion fallback, saturation or negative risk is allowed.

## Safety conclusion

- Safe edit boundary: preserve monotonic cumulative watermarks, exact order-key identity, five-bucket
  conservation and unchanged-state-on-error.
- High-risk impact: yes — this transfers HELD risk into actual FILLED usage for both KR and US.
