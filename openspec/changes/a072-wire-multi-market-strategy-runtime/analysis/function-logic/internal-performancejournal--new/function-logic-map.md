# Function Logic Map: `New`

- Source: `internal/performancejournal/adapter.go`
- Qualified function: `New`
- AST evidence: `ast.json` (base revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The constructor accepted only `*journal.ReadOnly`. The returned adapter stored that SELECT-only capability;
it did not accept a journal writer, performance store, broker or operating configuration.

## Branches and early returns

The base constructor is branchless and returned a `Reader` wrapping the exact read-only pointer. Nil was
allowed at construction but later read methods failed closed.

## Calls and live bindings

No calls or external bindings. The type-level `source` interface exposed one SELECT operation.

## State mutations and fallbacks

Allocates one in-memory adapter only. It grants no mutation capability and has no fallback data source.

## Safety conclusion

- Safe edit boundary: performance attribution consumes one narrow read-only projection.
- High-risk impact: low for execution, high for attribution correctness if broader capability leaks in.
