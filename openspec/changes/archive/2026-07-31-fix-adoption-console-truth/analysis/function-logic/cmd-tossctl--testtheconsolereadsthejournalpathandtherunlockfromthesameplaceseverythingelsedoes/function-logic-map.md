# Function Logic Map: `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| console assembly source | valid Go source containing reviewed resolver and run-lock calls | `cmd/tossctl/console.go` | test fails when either binding disappears or writable journal open appears |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | inspect the assembly source for the two path bindings | reads source only | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readSource` | load assembly source without executing it | test failure on read error | CodeGraph + AST |

## State mutations and fallbacks

- No runtime, file, broker, engine, or journal mutation.

## Safety conclusion

- Safe edit boundary: assertion follows the new active-profile resolver while preserving the read-only guard.
- High-risk impact: no; this is static regression evidence.
