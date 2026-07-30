# Function Logic Map: `parseRat`

- Source: `internal/exitpolicy/decimal.go`
- AST evidence: `ast.json` (`d800b958fc5aac6e4df8535baae722264606f23a0812d8784337f1b816e3e185`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 63 | only the branch's documented state transition | existing return/error contract | `TestparseRat` |
| B2 | existing if branch at line 66 | only the branch's documented state transition | existing return/error contract | `TestparseRat` |
| B3 | existing if branch at line 70 | only the branch's documented state transition | existing return/error contract | `TestparseRat` |

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
