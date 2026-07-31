# Function Logic Map: `ValidateRecoveryDerivation`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| exact line + immutable policy definition + remaining quantity | one validated policy arm, matching identity, positive remaining quantity | persisted a041 snapshot and a042 recovery evidence | `ErrRecoveryIdentity` or decimal validation error; never reconstructs a substitute |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | snapshot identity/shape invalid, policy arm count != 1, or quantity invalid | none | fail closed | recovery forgery table |
| B2 | ratchet identity/level/action valid | rederive from entry/risk/high/baseline | exact next line or refusal | existing exact derivation tests |
| B3 | ladder definition/identity/rung valid, including `NoRung=-1` | rederive via `nextLadderLine` | exact next line or refusal | `TestRecoveryAllowsLadderBeforeFirstRung` |
| B4 | derived line differs | none | `ErrRecoveryIdentity` | forged next-line tests |
| B5 | projection semantics invalid | none | `ErrRecoveryIdentity` | semantic output table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| policy identity functions | bind executable definition to persisted identity | errors propagate; no fallback registry lookup | CodeGraph + AST |
| `nextRatchetLine` / `nextLadderLine` | independently rederive next target/protection | exact error propagation | CodeGraph + AST |
| output/projection validators | enforce policy-kind and executable fields | fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Pure validation only; no state mutation, registry fallback, or execution authority.

## Safety conclusion

- Safe edit boundary: durable snapshot recovery validation.
- High-risk impact: yes; invalid recovery evidence must quarantine/refuse upstream.
