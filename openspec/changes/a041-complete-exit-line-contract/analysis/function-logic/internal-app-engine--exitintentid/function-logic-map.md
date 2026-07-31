# Function Logic Map: `exitIntentID`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| exit decision ID | `eld_` plus 64 hex characters | exit snapshot | invalid input returns empty fallback signal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | wrong prefix/length | none | empty | contract tests |
| B2 | valid decision ID | none | deterministic `exit_` key | concurrent engine test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| string normalization | remove surrounding whitespace only | no error | AST |

## State mutations and fallbacks

- The output links journal pending intent and mutation idempotency to DecisionID.

## Safety conclusion

- Safe edit boundary: pure identifier derivation.
- High-risk impact: yes — mutation dedup key.
