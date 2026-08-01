# Function Logic Map: `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| resolved profile paths | valid test/domain fixture | static source contract | fail test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | no production mutation | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `assert engine-owned commander wiring` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- no production mutation.

## Safety conclusion

- Safe edit boundary: not-applicable test evidence.
- High-risk impact: no.
