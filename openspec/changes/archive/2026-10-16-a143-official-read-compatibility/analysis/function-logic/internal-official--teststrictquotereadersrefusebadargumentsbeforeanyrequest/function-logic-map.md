# Function Logic Map: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest`

- Source: `internal/official/strict_quote_reads_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| invalid market/symbol cases | each must be refused | strict quote contract | test fails if either reader accepts or calls broker |
| `0101N0` | minute-only KR symbol | amended a143 scope | both quote readers must return `SYMBOL_INVALID`, zero requests |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate refusal cases | isolated offline harness | subtests run | this test |
| B2 | orderbook accepts invalid input | none | test fails | this test |
| B3 | orderbook returns refusal | validates typed error | proceeds | this test |
| B4 | alphanumeric case has wrong reason | none | test fails | this test |
| B5 | last-price accepts invalid input | none | test fails | this test |
| B6 | last-price returns refusal | validates typed error | proceeds | this test |
| B7 | alphanumeric case has wrong reason | none | test fails | this test |
| B8 | any broker request occurred | none | test fails | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrictOrderbookTop` | prove first quote reader refuses | offline harness; no request allowed | AST B2-B4 |
| `StrictLastPrice` | prove second quote reader refuses | offline harness; no request allowed | AST B5-B7 |
| `harness.hitCount` | prove pre-request refusal | must be zero | AST B8 |

## State mutations and fallbacks

- Test-only local HTTP harness. Production quote code is not modified.

## Safety conclusion

- Safe edit boundary: add the `0101N0` case and its exact reason checks while retaining the old letter case.
- High-risk impact: verifies read-side no-request policy; no order behavior changes.
