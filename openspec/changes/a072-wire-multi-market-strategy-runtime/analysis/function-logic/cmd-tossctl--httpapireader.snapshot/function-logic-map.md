# Function Logic Map: `httpAPIReader.Snapshot`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

One SSE full snapshot must contain the same resources as individual GET routes, including KR/US strategy runtime and bounded performance attribution.

## Branches and early returns

B1–B9 are the sequential resource reads and error returns; any failed resource prevents a mixed-generation full snapshot.

## Calls and live bindings

Calls each read-only resource method then encodes the shared DTO projections, including `StrategyRuntime` and `PerformanceFrom`.

## State mutations and fallbacks

No persistent state changes. Strategy runtime falls back only through its typed dormant projection owner, not through this snapshot combiner.

## Safety conclusion

SSE convergence gains data only; no command or mutation route is reachable.
