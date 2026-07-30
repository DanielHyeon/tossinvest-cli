# Function Logic Map: `NewExitObserver`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | non-nil journal/price/retrier/issuer/gateway, account, costs, valid policy snapshots | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing switch branch at line 250 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B2 | existing case branch at line 251 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B3 | existing case branch at line 253 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B4 | existing case branch at line 256 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B5 | existing case branch at line 259 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B6 | existing case branch at line 262 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B7 | existing case branch at line 264 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B8 | existing case branch at line 266 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B9 | existing if branch at line 272 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B10 | existing if branch at line 275 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B11 | existing if branch at line 279 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B12 | existing if branch at line 282 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B13 | existing if branch at line 285 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B14 | existing if branch at line 286 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B15 | existing if branch at line 293 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |
| B16 | existing if branch at line 296 | only the branch's documented state transition | existing return/error contract | `TestNewExitObserver` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| DefaultRatchetConfig, DefaultLadderPolicy, Validate, Clock.Now | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- reject unknown configured policy; do not expand the order or Guardian capabilities.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: reject unknown configured policy; do not expand the order or Guardian capabilities.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
