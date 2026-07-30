# Function Logic Map: `Journal.OpenAdoptedExitState`

- Source: `internal/journal/adoption.go`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | position and adoption record exist and no exit state exists | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 281 | only the branch's documented state transition | existing return/error contract | `TestOpenAdoptedExitStatePolicySnapshot` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| BeginTx, readAdoptionTx, OpenRatchetState, appendExitEventTx, Commit | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- use adoption.exit_policy_id in the same transaction and never re-read config.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: use adoption.exit_policy_id in the same transaction and never re-read config.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
