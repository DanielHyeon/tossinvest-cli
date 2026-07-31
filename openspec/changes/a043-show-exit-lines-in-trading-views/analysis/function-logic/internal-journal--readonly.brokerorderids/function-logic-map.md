# Function Logic Map: `ReadOnly.BrokerOrderIDs`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if branch at source line 298 | bounded test/read-model control flow only | ReadOnly.BrokerOrderIDs coverage and focused package suite |
| B2 | for branch at source line 304 | bounded test/read-model control flow only | ReadOnly.BrokerOrderIDs coverage and focused package suite |
| B3 | if branch at source line 306 | bounded test/read-model control flow only | ReadOnly.BrokerOrderIDs coverage and focused package suite |
| B4 | if branch at source line 311 | bounded test/read-model control flow only | ReadOnly.BrokerOrderIDs coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Existing read-only origin query is unchanged and re-mapped because an adjacent query changed the diff span. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Existing read-only origin query is unchanged and re-mapped because an adjacent query changed the diff span.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
