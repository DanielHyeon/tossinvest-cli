# Function Logic Map: `validateJudgementSnapshot`

- Source: `internal/journal/exit_snapshot.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement + sealed stored snapshot | exact position/provenance/observed state; proposal equals immutable executable action/level, or typed pre-arm suppression | a041 snapshot is execution handoff | caller wraps refusal as `ErrInvalidRequest` before transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | stored snapshot semantic validation fails | none | refuse | recovery forgery tests |
| B2 | flattened judgement fields differ | none | refuse | identity contract tests |
| B3 | nonorderable snapshot carries proposal/suppression | none | refuse | proposal coherence tests |
| B4 | orderable snapshot has nil proposal without known reason | none | refuse | missing/unknown suppression cases |
| B5 | known working-order suppression with nil proposal | preserve exact orderable snapshot | accept | typed arm suppression tests |
| B6 | proposal action/level differs or reason coexists | none | refuse | adversarial proposal table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StoredExitSnapshot.validate` | verify digest and semantic policy rederivation | error propagates | CodeGraph + AST |
| `ExecutableProposal` | obtain exactly the evaluator-approved proposal | no reconstruction from judgement fields | CodeGraph + AST |

## State mutations and fallbacks

- Pure pre-write validation. Saved-monotone selection may suppress a matching proposal later inside the transaction; this function validates the recomputed candidate first.

## Safety conclusion

- Safe edit boundary: immutable snapshot-to-arm coherence.
- High-risk impact: yes; prevents mismatched orders without dropping protective state updates.
