# Function Logic Map: `Client.ConditionalOrdersRaw`
- Source: `internal/official/conditional_reads.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Existing raw read retains its endpoint, cursor bounds, and verbatim decimal semantics; new client identity is additive and used by the separate protection adapter.
## Branches and early returns
- B1-B3 reject oversized query inputs; B4-B6 encode optional filters; B7 adapts each response order without deriving market or decimal values.
## Calls and live bindings
- Calls the official account GET transport only; errors propagate without WTS fallback.
## State mutations and fallbacks
- Builds a response slice only; no broker or local state changes.
## Safety conclusion
- Existing official read behavior is preserved while protection-specific adaptation remains isolated.
