# Function Logic Map: `Client.ListCompletedOrdersRange`

- Source: `internal/client/completed_orders.go`
- AST evidence: `ast.json` (`8ff2baf9106ab18ee8b721758f9354b886bf5b9b74df35193b265d1aa841a5ae`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| session / trading metadata | valid existing session and metadata | `requireSession`, `ensureTradingMetadata` | propagate error |
| market | `us`, `kr`, or `all` | `normalizeHistoryMarkets` | reject unsupported market |
| each response `result.body` | required JSON array; `[]` valid | a142 spec | missing/null/wrong type errors; never return partial aggregate |
| orders | existing parser/sort semantics | `parseCompletedOrder`, stable sort | preserve behavior |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | session failure | none | error | existing no-session coverage |
| B2 | metadata failure | read-only metadata discovery | error | existing metadata coverage |
| B3 | invalid market | none | error | existing normalization tests |
| B4 | each normalized market | one read-only GET per market | continue | valid US/KR fixtures |
| B5 | GET/decode error | none beyond attempted GET | error, discard accumulated orders | partial-market HTTP/wrong-type test |
| B6 | response body range | no mutation | append parsed orders | existing parse test |
| B7-B9 | sort comparator nil/non-nil time branches | local ordering only | sorted output | existing ordering tests |
| planned guard | `result.body == nil` after a successful decode | none | descriptive schema error, discard aggregate | missing/null body tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireSession`, `ensureTradingMetadata` | authenticated WTS access | propagated | AST B1-B2; CodeGraph |
| `normalizeHistoryMarkets` | determine required markets | propagated | AST B3 |
| `getJSON` | per-market completed-order GET | propagated; no partial successful return | AST B5 |
| `parseCompletedOrder`, `sort.SliceStable` | preserve current mapping/order | no remote mutation | AST B6-B9 |

## State mutations and fallbacks

- Local aggregate only. It must be abandoned on every required-market schema/transport failure.
- No order, ledger, reconciliation, or live setting mutation.

## Safety conclusion

- Safe edit boundary: validate `envelope.Result.Body` immediately after each successful decode and before append.
- High-risk impact: yes. Completion lookup and delayed-cancel recovery consume this data; false absence is forbidden.
