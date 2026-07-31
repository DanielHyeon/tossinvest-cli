# Function Logic Map: `ValidateRecoveryDerivation`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| exact line + exact evaluator input | one validated policy arm containing the original snapshot context, policy input, quantity and prior watermark/protection/stage inside that input | persisted a041 snapshot and a042 recovery evidence | `ErrRecoveryIdentity`; never reconstructs a substitute from output fields |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | snapshot identity/shape invalid or policy arm count != 1 | none | fail closed | recovery forgery table |
| B2 | ratchet arm | rerun `EvaluateRatchetSnapshot` then `ChangedFromState` | exact full line or refusal | ratchet recovery/forged-level tests |
| B3 | ladder arm, including `NoRung=-1` | rerun `EvaluateLadderSnapshot` then `ChangedFromState` | exact full line or refusal | ladder recovery tests |
| B4 | evaluator error | none | wrapped `ErrRecoveryIdentity` | invalid input tests |
| B5 | any derived field differs | none | `ErrRecoveryIdentity` | quantity/protection/level/cancel/changed forgery table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EvaluateRatchetSnapshot` / `EvaluateLadderSnapshot` | rerun the same pure evaluator over persisted exact inputs | error wraps as identity refusal; no fallback registry lookup | CodeGraph + AST |
| `ChangedFromState` | bind the evaluator input's original watermark/protection/stage and the `Changed` bit | exact value comparison; no redundant wrapper fields | CodeGraph + AST |

## State mutations and fallbacks

- Pure validation only; `reflect.DeepEqual` covers every line field after exact re-evaluation. No state mutation, registry fallback, or execution authority.

## Safety conclusion

- Safe edit boundary: durable snapshot recovery validation.
- High-risk impact: yes; invalid recovery evidence must quarantine/refuse upstream.
