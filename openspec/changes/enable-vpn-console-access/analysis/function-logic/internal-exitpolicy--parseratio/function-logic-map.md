# Function Logic Map: `parseRatio`

- Source: `internal/exitpolicy/decimal.go`
- AST evidence: `ast.json` (`eded0b568d7c050b7249fb7ca92cebca1cb1cde5949540f2b7b490b740d77ca4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 82 | only the branch's documented state transition | existing return/error contract | `TestparseRatio` |
| B2 | existing if branch at line 86 | only the branch's documented state transition | existing return/error contract | `TestparseRatio` |
| B3 | existing if branch at line 91 | only the branch's documented state transition | existing return/error contract | `TestparseRatio` |

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
