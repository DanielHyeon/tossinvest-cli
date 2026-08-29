# Function Logic Map: `scanAdoption`

- Source: `internal/journal/adoption.go`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | query column order exactly matches adoption fields | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 348 | only the branch's documented state transition | existing return/error contract | `TestPositionAdoptionPolicyID` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| row.Scan | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- read nullable exit_policy_id as empty legacy value.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: read nullable exit_policy_id as empty legacy value.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
