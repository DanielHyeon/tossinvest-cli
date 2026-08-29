# Function Logic Map: `fakePerformanceReader.Dashboard`

- Source: `internal/console/performance_history_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

This test fake records the exact server query and returns fixture data.

## Branches and early returns

Branchless B1 covers its deterministic return.

## Calls and live bindings

No production callee or live binding exists.

## State mutations and fallbacks

Only test counters and fixture query fields change.

## Safety conclusion

Test-only; no production impact.
