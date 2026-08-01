# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement scope | position plus lifecycle generation | ExitObserver state | invalid/stale refused before mutation |
| snapshot/proposal | coherent immutable tuple | exitpolicy evaluator | validation error, no partial write |
| current exit row | incomplete and same generation | journal DB | completed/stale generation refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid id/provenance/proposal/snapshot | none | validation error | exit-state/snapshot tests |
| B2 | transaction/read fails | none | wrapped error | journal tests |
| B3 | completed state | none | `ErrExitStateCompleted` | exit-state tests |
| B4 | lifecycle generation mismatch/released | none | version mismatch/refusal | late-generation test |
| B5 | duplicate decision | none | proposal pending | identity tests |
| B6 | monotonicity or policy mismatch | none | refusal | snapshot tests |
| B7 | valid judgement | update state, optional arm, append event, commit | result | exit-loop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitProgress` | lock/read current state | fail closed | CodeGraph + AST |
| `appendExitEventTx` | append generation-bound history | same transaction | CodeGraph + AST |
| `armExitProposalTx` | durable pre-submit arm | same transaction | CodeGraph + AST |

## State mutations and fallbacks

- One write transaction; every refusal occurs before commit. Generation comparison must precede monotone recovery and arming.

## Safety conclusion

- Safe edit boundary: add exact lifecycle-generation/status guard before any update/event/arm.
- High-risk impact: yes — sole judgement persistence authority.
