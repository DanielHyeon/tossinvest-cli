# Function Logic Map: `PerformanceFrom`

- Source: `internal/httpapi/read.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Projection preserves fixed query, completeness counts, aggregates and every exact attribution field without recalculation.

## Branches and early returns

B1–B5 cover optional newest-source time, aggregate iteration, metric iteration, attribution iteration and close-leg iteration.

## Calls and live bindings

Only pure DTO helpers `performancePnLFrom` and `performanceAmountFrom` are called.

## State mutations and fallbacks

Allocates output slices and UTC timestamp copies. Missing values remain empty with explicit status.

## Safety conclusion

Pure read projection; no persistence or trading capability.
