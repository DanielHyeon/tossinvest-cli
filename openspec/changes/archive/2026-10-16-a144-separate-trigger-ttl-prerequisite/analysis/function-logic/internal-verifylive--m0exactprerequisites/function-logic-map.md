# Function Logic Map: `M0ExactPrerequisites`

- Source: `internal/verifylive/m0_recovery.go`
- Base: `c564a89bf8fbea1a3c29e5c0b5952f8cddd3ef34`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `entries` | prior verification record | `LastEntry`/`Passed` | missing or non-PASS required step returns HOLD error |
| `Steps()` | procedure catalogue | `internal/verifylive/verifylive.go` | every current/future step is checked unless fixed exception applies |
| M0 exception | exactly trigger and TTL-edge IDs | predicate | TTL does not become PASS; all other steps remain required |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over `Steps()` | none | evaluates all catalogue entries | `TestM0ExactPrerequisitesStillRejectsEveryNonTTLNonPass` |
| B2 | ID is `conditional-trigger` or `idempotency-ttl-edge` | none | skip prerequisite test | `TestM0TriggerAdmitsEveryTTLVerdictWithoutTTLMutation`; `TestM0CLIValidationTreatsOnlyTTLAsNonPrerequisite` |
| B3 | any remaining step lacks latest PASS | none | error naming rejected step | `TestM0ExactPrerequisitesStillRejectsEveryNonTTLNonPass` |
| return | all remaining steps PASS | none | nil | same TTL and CLI tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Steps` | enumerates fixed procedure requirements | none | AST B1 |
| `Passed` | reads latest recorded verdict | false on missing/non-PASS | AST B3 |
| `fmt.Errorf` | reports local admission HOLD | no broker call | AST B3 |

## State mutations and fallbacks

- No mutation, broker call, record write, setting, or approval interaction occurs in this predicate.
- Caller order remains unchanged: CLI/direct Runner applies outstanding and M0 checkpoint holds before this predicate; `pending-create` keeps its read-only recovery path.
- Missing/nonterminal TTL is handled later by existing no-opt-in preflight as ordinary `skipped`; terminal TTL lines stay settled.

## Safety conclusion

- Safe edit boundary: one fixed ID predicate in `M0ExactPrerequisites`.
- High-risk impact: trigger admission governs an irreversible measurement. Test coverage asserts zero TTL planned mutation and zero TTL order POST/DELETE while retaining every non-TTL failure hold.
