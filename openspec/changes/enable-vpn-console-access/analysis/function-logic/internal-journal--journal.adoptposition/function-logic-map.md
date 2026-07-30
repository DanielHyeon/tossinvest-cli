# Function Logic Map: `Journal.AdoptPosition`

- Source: `internal/journal/adoption.go`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | position has neither entry decision nor prior different adoption | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 180 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B2 | existing if branch at line 185 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B3 | existing if branch at line 199 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B4 | existing if branch at line 202 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B5 | existing if branch at line 205 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B6 | existing if branch at line 208 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B7 | existing if branch at line 213 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B8 | existing if branch at line 220 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B9 | existing if branch at line 229 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B10 | existing if branch at line 243 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B11 | existing if branch at line 248 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B12 | existing if branch at line 252 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |
| B13 | existing if branch at line 256 | only the branch's documented state transition | existing return/error contract | `TestAdoptPositionPolicySnapshot` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| record, BeginTx, readAdoptionTx, ExecContext, Commit | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- persist selected policy ID with adoption before setting the position pointer.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: persist selected policy ID with adoption before setting the position pointer.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
