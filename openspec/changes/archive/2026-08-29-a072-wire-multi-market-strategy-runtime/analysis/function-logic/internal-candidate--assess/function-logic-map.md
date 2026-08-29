# Function Logic Map: `Assess`

- Source: `internal/candidate/watch.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Assessment requires KR or US, a non-zero observation time and a validated threshold set. It may read only the requested market.

## Branches and early returns

B1–B7 cover invalid store/options, market normalization, threshold errors, market-scoped query failure, candidate iteration, expired exclusion and readiness aggregation.

## Calls and live bindings

`summariesForMarket` pushes `WHERE market=?` into SQLite; lifecycle and threshold evaluation remain unchanged.

## State mutations and fallbacks

Assessment is read-only and never copies peer-market candidates.

## Safety conclusion

The optimization reduces work without changing candidate authority semantics.
