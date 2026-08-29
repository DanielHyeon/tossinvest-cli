# Function Logic Map: `Collector.holdings`

- Source: `internal/reconcile/snapshot.go`
- AST evidence: `ast.json` — lines 315-350, branches 5
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Collector.Positions` | raw-string reader or legacy float reader | configured broker adapter | adapter error is returned unchanged |
| `RawPosition.Quantity` | blank, exact finite decimal, or unreadable broker evidence | official account snapshot | raw path preserves unreadable evidence so comparison fails closed |
| legacy `Position.Quantity` | finite `float64` | legacy reader | converted with the established decimal formatter |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reader implements `RawPositionsReader` | select lossless raw path | continue or return adapter error | `TestA110CollectorPreservesUnreadableRawHoldingQuantity` |
| B2 | raw adapter returns error | none | return error | collector adapter-error regression |
| B3 | each raw position | append normalized holding; quantity uses `canonicalHoldingQuantity` | return raw holdings | `TestA110CollectorPreservesUnreadableRawHoldingQuantity` |
| B4 | legacy adapter returns error | none | return error | collector adapter-error regression |
| B5 | each legacy position | append established float-formatted holding | return legacy holdings | collector snapshot regressions |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RawPositionsReader.PositionsRaw` | obtain quantity without a float conversion | propagate errors; no fallback to lossy legacy path | AST B1-B3 |
| `canonicalHoldingQuantity` | exact-canonicalize finite quantity while retaining blank/malformed/non-finite text | unreadable input is returned trimmed, never rewritten to zero | AST B3; A110 collector RED/GREEN |
| `PositionsReader.Positions` | compatibility path for already-typed floats | propagate errors | AST B4-B5 |
| `canonicalDecimal` | normalize average price, whose blank-as-zero digest contract remains unchanged | unreadable text remains visible | AST B3 |

## State mutations and fallbacks

- The function only constructs and returns a new holdings slice; it does not mutate engine or journal state.
- The raw and legacy reader paths are mutually exclusive. A raw-reader failure is not retried through the lossy legacy path.
- Quantity is the safety boundary: raw unreadable evidence survives until `Comparer.Compare`, which emits an ordinary fail-closed mismatch that cannot earn permanent promotion.

## Safety conclusion

- Safe edit boundary: use the dedicated quantity helper only for `RawPosition.Quantity`; do not change global blank-as-zero digest behavior.
- High-risk impact: yes. Rewriting blank evidence to zero can erase a mismatch and suppress the reconcile/adoption gate.
