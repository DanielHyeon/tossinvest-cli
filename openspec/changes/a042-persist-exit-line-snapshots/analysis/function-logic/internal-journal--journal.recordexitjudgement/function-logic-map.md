# Function Logic Map: `Journal.RecordExitJudgement`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement snapshot/provenance/proposal | one position generation, immutable policy tuple, coherent snapshot; proposal provenance exact-match | a041 snapshot, a042 recovery spec | reject or durably quarantine without state/proposal mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | validate position/provenance/snapshot/proposal tuple | none | typed refusal | provenance and malformed tuple tests |
| B8-B12 | transaction/read current state/dedup/completion/monotone checks | read current effective | no-op duplicate or rollback | concurrent decision and monotone tests |
| B13-B17 | select whole saved/recomputed candidate; update effective; arm; append event; commit | one atomic write | rollback on any fault | atomic fault matrix and crash reopen tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `selectRecoverySnapshot` | select one coherent tuple, never field-wise max | ambiguity returns quarantine cause | CodeGraph + AST |
| `armExitProposalTx`, `appendExitEventTx` | preserve arm-before-submit and evaluation history | same transaction; no network call | CodeGraph + AST |

## State mutations and fallbacks

- State, proposal arm, and evaluation event remain inside one `BEGIN IMMEDIATE` transaction.
- Duplicate decision IDs are idempotent; semantic ambiguity is quarantined by exact position generation.

## Safety conclusion

- Safe edit boundary: journal-only atomic write; broker submission remains after commit in engine.
- High-risk impact: yes; crash and race tests required.
