# Function Logic Map: `Client.ListPendingOrders`

- Source: `internal/client/account.go`
- AST evidence: `ast.json` (`cda0d5d3cb12e91e55641aa66cc54f02f7e77ed48b57a4052b83c09d39b0a90f`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Failure behavior |
|---|---|---|
| session | stored cookies exist | propagated error |
| response `result` | JSON array; `[]` valid | absent/null errors; incompatible JSON errors |
| items | existing parser contract | retain parse behavior |

## Branches and early returns

| Branch | Condition | Return/error | Test |
|---|---|---|---|
| B1 | session missing | error | existing client coverage |
| B2 | GET/decode failure | propagated error | `TestListPendingOrdersRejectsAbsentOrNullResult` |
| B3 | decoded result is nil | schema error | `TestListPendingOrdersRejectsAbsentOrNullResult` |
| B4 | collection present | parsed results, including empty non-nil slice | `TestListPendingOrdersAcceptsExplicitEmptyResult`; `TestListPendingOrdersParsesStructuredFields` |

## Safety conclusion

Read-only only. The post-decode guard prevents missing or null data from becoming an empty-order claim.

## Calls and live bindings
Calls are listed in `ast.json`; fixture tests use only local `httptest`.

## State mutations and fallbacks
No mutation or fallback is introduced by this evidence.
