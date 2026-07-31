# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement snapshot/provenance/proposal and current durable state | exact recovery input/output; one coherent generation; legacy carries no suppression claim | a041/a042 contracts | refusal, rollback, or generation quarantine |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B43 | validation, transaction, dedup, recovery selection, write hooks, arm/event/commit, result projection | one local atomic transaction | typed error or committed result | journal recovery, fault, crash and concurrency suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SelectRecoverySnapshot` / `armExitProposalTx` / `appendExitEventTx` | coherent candidate selection and atomic state/arm/audit persistence | no network call; every error rolls back | CodeGraph + AST |

## State mutations and fallbacks

- Saved-monotone selection clears request proposal and arm-suppression reason before persistence.
- The output pointer is populated only after successful commit and contains a proposal only for `ExitArmArmed`.

## Safety conclusion

- Safe edit boundary: journal transaction; broker submission remains outside.
- High-risk impact: yes; crash/race/saved-monotone tests are mandatory.
