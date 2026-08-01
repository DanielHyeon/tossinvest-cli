# Function Logic Map: `attachPositionExitLines`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range branch at source line 59 | bounded test/read-model control flow only | attachPositionExitLines coverage and focused package suite |
| B2 | if branch at source line 61 | bounded test/read-model control flow only | attachPositionExitLines coverage and focused package suite |
| B3 | if branch at source line 71 | bounded test/read-model control flow only | attachPositionExitLines coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Purely select persisted snapshots, apply the approved display freshness bound, and build canonical views without policy evaluation. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Purely select persisted snapshots, apply the approved display freshness bound, and build canonical views without policy evaluation.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
