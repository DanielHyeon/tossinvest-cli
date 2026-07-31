# Function Logic Map: `ReadOnly.BrokerOrderExitLinks`

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
| B1 | if branch at source line 332 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B2 | for branch at source line 343 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B3 | if branch at source line 345 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B4 | if branch at source line 352 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B5 | if branch at source line 356 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B6 | range branch at source line 363 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B7 | if branch at source line 364 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B8 | if branch at source line 365 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |
| B9 | if branch at source line 372 | bounded test/read-model control flow only | ReadOnly.BrokerOrderExitLinks coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Exact SQL lineage joins broker order attempts to exit events only through the persisted intent ID and fails duplicate-event evidence closed. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Exact SQL lineage joins broker order attempts to exit events only through the persisted intent ID and fails duplicate-event evidence closed.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
