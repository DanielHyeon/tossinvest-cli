# Function Logic Map: `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range branch at source line 284 | bounded test/read-model control flow only | TestThePositionsScreenShowsTheExitLineOfAManagedPosition coverage and focused package suite |
| B2 | if branch at source line 290 | bounded test/read-model control flow only | TestThePositionsScreenShowsTheExitLineOfAManagedPosition coverage and focused package suite |
| B3 | range branch at source line 295 | bounded test/read-model control flow only | TestThePositionsScreenShowsTheExitLineOfAManagedPosition coverage and focused package suite |
| B4 | if branch at source line 296 | bounded test/read-model control flow only | TestThePositionsScreenShowsTheExitLineOfAManagedPosition coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Legacy fixture now asserts managed identity is retained while unsealed raw exit values remain hidden. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Legacy fixture now asserts managed identity is retained while unsealed raw exit values remain hidden.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
