# Function Logic Map: `evaluate`

- Source: `internal/reversallane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed reversal authority and explicit entry/target terms. Effective stop is the exact non-retreat candidate selected by current stop rules.

## Branches and early returns

B1-B16 cover OFF, invalidation, plan/FX/schema/risk/stop/leg/cap/structure/quantity/admission. A new term branch rejects missing or unordered prices.

## Calls and live bindings

Pure metric, structure, allocation and risk functions only.

## State mutations and fallbacks

None; accepted output preserves canonical terms, refusal has none.

## Safety conclusion

Safe edit boundary: canonicalize the already validated non-retreat stop and require `stop < entry < target`. High-risk impact: no inferred target.
