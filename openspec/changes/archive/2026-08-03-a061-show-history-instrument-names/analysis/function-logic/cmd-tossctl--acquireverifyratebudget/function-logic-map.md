# Function Logic Map: `acquireVerifyRateBudget`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | live CLI/console verification context | command or run state | cancellation stops waiting before broker construction |
| active profile | default data dir or isolated `--config-dir` | `engineJournalDir(root)` | path error returns before any lease or API call |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | active profile budget path cannot be derived | none | wrapped path error | profile path tests |
| B2 | rate-budget lease cannot be acquired before cancellation/error | no broker or runner construction | wrapped error | `TestA061VerificationWaitsForAndThenOwnsTheSameRateBudget` |
| tail | lease acquired | kernel holds profile lock and diagnostic is printed | held lease | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ratebudget.Acquire` | exclude optional metadata across processes | waits in cancellable 25ms polls | ratebudget unit tests |
| `verifyRateBudgetPath` | bind marker and lease to the active profile, independent of `--record` | pure path derivation | override isolation test |

## State mutations and fallbacks

- The kernel lease is the only state; callers defer release for the full verification lifetime.

## Safety conclusion

- Safe edit boundary: admission before any broker call; no order or account mutation.
- High-risk impact: yes, because verification must own the budget before it constructs the live broker.
