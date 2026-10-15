# Function Logic Map: `Client.ListCompletedOrdersRange`

- Source: `internal/client/completed_orders.go`
- AST evidence: `ast.json` (`2c389d043a94552ec5dd4e0179f3f5c3cdbd1eb26eaefec8cbbfcae5959cc486`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Failure behavior |
|---|---|---|
| session and metadata | existing authenticated WTS contract | propagated error |
| market | `us`, `kr`, or `all` | invalid market errors |
| each `result.body` | JSON array; `[]` valid | absent/null errors; incompatible JSON errors; no partial aggregate |
| order mapping/order | existing parser and stable sort | preserve behavior |

## Branches and early returns

| Branch | Condition | Return/error | Test |
|---|---|---|---|
| B1 | session failure | error | existing client coverage |
| B2 | metadata failure | error | existing client coverage |
| B3 | invalid market | error | existing normalization coverage |
| B4 | normalized market iteration | one GET per market | partial-market fixture |
| B5 | GET/decode error | error; discard aggregate | `TestListCompletedOrdersRangeRejectsMissingCollections` |
| B6 | decoded body is nil | schema error; discard aggregate | `TestListCompletedOrdersRangeRejectsMissingCollections`; `TestListCompletedOrdersRangeRejectsPartialMarketResult` |
| B7 | body range | append parsed orders | parse and empty-body tests |
| B8 | both sort timestamps nil | ID ordering | existing ordering coverage |
| B9 | left timestamp nil | timestamped order first | existing ordering coverage |
| B10 | right timestamp nil | timestamped order first | existing ordering coverage |

## Safety conclusion

Read-only aggregate. Every successful decode now proves its required collection exists before any append.

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.
