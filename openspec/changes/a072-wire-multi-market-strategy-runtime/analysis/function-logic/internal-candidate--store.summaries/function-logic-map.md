# Function Logic Map: `Store.Summaries`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The public all-market summary preserves its contract and delegates to the same query helper with no market predicate.

## Branches and early returns

The function is branchless; B1 is the helper success/error pass-through.

## Calls and live bindings

Calls `summariesForMarket(ctx, "", at)`, which performs immutable candidate reads and canonical lifecycle projection.

## State mutations and fallbacks

No writes or defaults are introduced.

## Safety conclusion

Existing callers retain exact behavior while production assessment can use a separate market-scoped helper.
