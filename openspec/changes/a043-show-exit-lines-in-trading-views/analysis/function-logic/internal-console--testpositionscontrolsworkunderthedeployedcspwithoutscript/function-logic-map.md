# Function Logic Map: `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range branch at source line 25 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B2 | if branch at source line 30 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B3 | range branch at source line 34 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B4 | if branch at source line 36 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B5 | range branch at source line 40 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B6 | range branch at source line 43 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B7 | if branch at source line 44 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |
| B8 | if branch at source line 49 | bounded test/read-model control flow only | TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Both trading views reject POST and contain no form, input, textarea, select, button, or contenteditable surface. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Both trading views reject POST and contain no form, input, textarea, select, button, or contenteditable surface.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
