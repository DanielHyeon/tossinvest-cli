# Function Logic Map: `TestListPendingOrdersParsesStructuredFields`

- Source: `internal/client/orders_parse_test.go`
- AST evidence: `ast.json` (`84a0f5c3e3342b54907c3db6b66d2540acd451ecd909ec6c257429979c1f903d`, base revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Local fixture test baseline.

## Branches and early returns

| Branch | Baseline condition |
|---|---|
| B1 | request path mismatch |
| B2 | read error |
| B3 | count != 1 |
| B4 | order ID mismatch |
| B5 | symbol mismatch |
| B6 | price mismatch |
| B7 | status mismatch |


## Calls and live bindings

Fixture-only calls; no live service.

## State mutations and fallbacks

No production mutation.

## Safety conclusion

Historical AST from the immutable baseline; file hash is intentionally the baseline hash. The test body remains unchanged; adjacent tests changed its source file. Named current test PASS does not mean fatal assertion branches execute.
