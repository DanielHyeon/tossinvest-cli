# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go`
- AST evidence: `ast.json` (`fb8a9644a78693dcae49774a9478cc2c400ad0c38ab7cf5793732c6b24400c9d`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | Context dependencies and caller overrides are validated | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 314 | only the branch's documented state transition | existing return/error contract | `TestContextExitObserverPolicy` |
| B2 | existing if branch at line 317 | only the branch's documented state transition | existing return/error contract | `TestContextExitObserverPolicy` |
| B3 | existing if branch at line 321 | only the branch's documented state transition | existing return/error contract | `TestContextExitObserverPolicy` |
| B4 | existing if branch at line 332 | only the branch's documented state transition | existing return/error contract | `TestContextExitObserverPolicy` |
| B5 | existing if branch at line 335 | only the branch's documented state transition | existing return/error contract | `TestContextExitObserverPolicy` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| NewExitObserver | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- inject config common policy only into observer construction.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: inject config common policy only into observer construction.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
