# Function Logic Map: `TestListCompletedOrdersParsesStructuredFields`

- Source: `internal/client/completed_orders_test.go`
- AST evidence: `ast.json` (`5dc5f6bd8966e6c957bd4fe85925a9a357a6a96ab4341ce78cb2d95f3d5d890f`, base revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Local fixture test baseline.

## Branches and early returns

| Branch | Baseline condition |
|---|---|
| B1 | request path mismatch |
| B2 | range.from mismatch |
| B3 | range.to mismatch |
| B4 | read error |
| B5 | count != 1 |
| B6 | order ID mismatch |
| B7 | symbol mismatch |
| B8 | market mismatch |
| B9 | status mismatch |


## Calls and live bindings

Fixture-only calls; no live service.

## State mutations and fallbacks

No production mutation.

## Safety conclusion

Historical AST from the immutable baseline; file hash is intentionally the baseline hash. The test body remains unchanged; adjacent tests changed its source file. Named current test PASS does not mean fatal assertion branches execute.
