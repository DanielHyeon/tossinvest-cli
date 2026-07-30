# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json` (`941f908df3758d1efc70be01dd276e47267fae61a5dcf8f8cfa1d2ba45ee924f`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | position exists and is exit-eligible; kind/ID are validated; t0 prices are valid | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 108 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B2 | existing if branch at line 112 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B3 | existing if branch at line 115 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B4 | existing switch branch at line 120 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B5 | existing case branch at line 121 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B6 | existing if branch at line 122 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B7 | existing case branch at line 125 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B8 | existing if branch at line 126 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B9 | existing else branch at line 128 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B10 | existing if branch at line 128 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B11 | existing if branch at line 136 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B12 | existing if branch at line 142 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B13 | existing if branch at line 151 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B14 | existing if branch at line 154 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B15 | existing if branch at line 157 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B16 | existing if branch at line 161 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B17 | existing if branch at line 168 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B18 | existing if branch at line 173 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |
| B19 | existing if branch at line 179 | only the branch's documented state transition | existing return/error contract | `TestOpenExitStatePolicySnapshot` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| OpenRatchetState, BeginTx, appendExitEventTx, Commit, ExitState | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- write policy ID in the same transaction as state open; unique/existence failures remain fail-closed.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: write policy ID in the same transaction as state open; unique/existence failures remain fail-closed.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
