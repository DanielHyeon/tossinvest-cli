# Function Logic Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go`
- AST evidence: `ast.json` (`66c4f4356e33a53bca02fa80c4c064058e4d03574d89138fcd85672fd07e8e40`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated decimal policy/state, positive entry/observed/high-water/baseline, monotone baseline | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 299 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B2 | existing if branch at line 302 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B3 | existing if branch at line 309 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B4 | existing if branch at line 313 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B5 | existing if branch at line 317 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B6 | existing if branch at line 321 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B7 | existing if branch at line 324 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B8 | existing if branch at line 327 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B9 | existing if branch at line 334 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B10 | existing range branch at line 353 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B11 | existing if branch at line 355 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B12 | existing if branch at line 358 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B13 | existing if branch at line 365 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B14 | existing if branch at line 367 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B15 | existing if branch at line 371 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B16 | existing if branch at line 373 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B17 | existing if branch at line 383 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B18 | existing if branch at line 387 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B19 | existing if branch at line 399 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B20 | existing if branch at line 406 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B21 | existing if branch at line 415 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B22 | existing if branch at line 418 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B23 | existing if branch at line 420 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B24 | existing if branch at line 432 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B25 | existing switch branch at line 436 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B26 | existing case branch at line 437 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B27 | existing case branch at line 440 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B28 | existing case branch at line 445 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |
| B29 | existing if branch at line 454 | only the branch's documented state transition | existing return/error contract | `TestEvaluateLadder` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Validate, positive/fraction/parseRat, ComputeProtectedStop, lockPrice | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- preserve promotion-before-breach, pending cancel-first, completion, and partial/final precedence; add runner only as a max-composed protection candidate.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: preserve promotion-before-breach, pending cancel-first, completion, and partial/final precedence; add runner only as a max-composed protection candidate.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
