# Function Logic Map: `LadderPolicy.Validate`

- Source: `internal/exitpolicy/ladder.go`
- AST evidence: `ast.json` (`66c4f4356e33a53bca02fa80c4c064058e4d03574d89138fcd85672fd07e8e40`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 167 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B2 | existing if branch at line 170 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B3 | existing range branch at line 173 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B4 | existing if branch at line 174 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B5 | existing if branch at line 177 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B6 | existing if branch at line 183 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B7 | existing if branch at line 189 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B8 | existing if branch at line 194 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B9 | existing if branch at line 196 | only the branch's documented state transition | existing return/error contract | `TestValidate` |
| B10 | existing if branch at line 199 | only the branch's documented state transition | existing return/error contract | `TestValidate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST-listed callees | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- preserve existing fail-closed behavior.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: preserve existing fail-closed behavior.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
