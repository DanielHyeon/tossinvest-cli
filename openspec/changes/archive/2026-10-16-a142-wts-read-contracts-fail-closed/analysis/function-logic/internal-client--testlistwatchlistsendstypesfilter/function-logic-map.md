# Function Logic Map: `TestListWatchlistSendsTypesFilter`

- Source: `internal/client/watchlist_test.go`
- AST evidence: `ast.json` (`3fe035beb0477413942969df4024973e4e4dae07467ed7c3fe48d3f54b2809c9`, base revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Local fixture test baseline.

## Branches and early returns

| Branch | Baseline condition |
|---|---|
| B1 | legacy dashboard path selects body capture |
| B2 | legacy read error |
| B3 | legacy types/WATCHLIST body assertion fails |


## Calls and live bindings

Fixture-only calls; no live service.

## State mutations and fallbacks

No production mutation.

## Safety conclusion

Historical AST from the immutable baseline; file hash is intentionally the baseline hash. This test was removed with the legacy route; current tests have separate bundles.
