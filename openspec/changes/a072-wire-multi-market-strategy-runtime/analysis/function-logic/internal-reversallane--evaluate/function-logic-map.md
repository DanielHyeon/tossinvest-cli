# Function Logic Map: `evaluate`

- Source: `internal/reversallane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed reversal authority and explicit entry/target terms. Admitted evaluation additionally
consumes the sealed a066 cap; proposal evaluation has no cap and emits immutable planned remaining as
`q_candidate`. Effective stop is the exact non-retreat candidate selected by current stop rules.

## Branches and early returns

B1-B9 cover OFF, invalidation, plan/FX/schema/risk/stop/leg. The shared evaluator then validates
final-leg structure before quantity authority, computes planned remaining, and branches explicitly:
proposal returns that remaining without a cap; admitted evaluation validates a066 cap, applies
`min(remaining,q_final)`, requires exact reservation quantity and performs risk admission. The term
branch rejects missing or unordered prices in both modes.

## Calls and live bindings

Pure metric, structure, allocation and risk functions only.

## State mutations and fallbacks

None; accepted proposal has no cap snapshot/policy lineage, while admitted output preserves it.
Both modes preserve canonical terms; refusal has none.

## Safety conclusion

Safe edit boundary: one shared evaluator with explicit proposal/admitted mode; only cap validation,
reservation equality and proposed-risk admission differ. High-risk impact: no inferred target and no
fabricated cap.
