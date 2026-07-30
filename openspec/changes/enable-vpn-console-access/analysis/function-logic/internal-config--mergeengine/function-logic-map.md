# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json` (`01f2158931852abd45c063f40ba7d9c6d9a346e28a1d8128daf4a6b3b8126a13`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | raw optional blocks merge onto safe zero-value config | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 299 | only the branch's documented state transition | existing return/error contract | `TestMergeExitPolicy` |
| B2 | existing if branch at line 303 | only the branch's documented state transition | existing return/error contract | `TestMergeExitPolicy` |
| B3 | existing if branch at line 308 | only the branch's documented state transition | existing return/error contract | `TestMergeExitPolicy` |
| B4 | existing if branch at line 312 | only the branch's documented state transition | existing return/error contract | `TestMergeExitPolicy` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| adoption/gate validation and normalization | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- empty means legacy RATCHET; unknown non-empty policy is retained as rejected and cannot run.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: empty means legacy RATCHET; unknown non-empty policy is retained as rejected and cannot run.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
