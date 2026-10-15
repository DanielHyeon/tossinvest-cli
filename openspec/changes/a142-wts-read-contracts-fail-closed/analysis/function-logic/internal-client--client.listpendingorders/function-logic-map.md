# Function Logic Map: `Client.ListPendingOrders`

- Source: `internal/client/account.go`
- AST evidence: `ast.json` (`0c8f47e783ee5bf16b47ae6ad68bc3f7f43f2617c5a4cc7e7871470f51e8b16e`)
- Risk scan: `risk-pattern-report.md` (no configured match)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| session | stored cookies exist | `requireSession` | return `ErrNoSession` |
| pending response `result` | required JSON array; `[]` valid | a142 spec | missing/null/wrong type returns error; never infer no orders |
| individual items | existing parser contract | `parsePendingOrder` | retain current best-effort item parsing scope |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `requireSession` fails | none | error | existing no-session coverage |
| B2 | GET/decode fails | one read-only GET | propagated error | wrong-type result fixture |
| B3 | range result items | no mutation | parsed result | existing parsing fixture |
| planned guard | decoded `result == nil` | none | descriptive schema error | missing/null result; `[]` success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireSession` | prevent anonymous WTS call | propagated | AST B1, CodeGraph callers/callees |
| `getJSON` | authenticated pending-order GET | status/transport/decode propagated; no retry here | AST B2 |
| `parsePendingOrder` | map valid raw item | existing compatibility only | AST B3 |

## State mutations and fallbacks

- No local or remote mutation. The GET result is returned only after collection validation.
- No fallback from schema failure to empty data or completed-history lookup.

## Safety conclusion

- Safe edit boundary: one post-decode collection-presence guard before allocation/range.
- High-risk impact: yes. Callers include order lookup/recovery and pending-order trading flows; change only preserves error propagation.
