# Function Logic Map: `TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated`

- Source: `internal/weeklyvaluelane/evaluate_registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request saved stop | scalar and private authority must model the same persisted state | test fixture | clear both to model no saved stop |
| structural invalidation | independent of entry admission and common exit | evaluator contract | invalidation outcome remains zero-quantity |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | sealed saved stop 90 beats candidate 80 | none | decision with stop 90 | this test |
| B2 | both saved scalar and authority cleared | none | structural cap refusal for candidate 80 | this test |
| B3 | structural invalidation | none | invalidation; common exit remains independent | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validEvaluation` | construct sealed baseline request | test failure, no retry | AST |
| `mintSavedStopAuthority` | bind saved stop to plan/evidence | zero authority on invalid input | AST |
| `EvaluateKR` | exercise decision/refusal/invalidation paths | typed outcome, no retry | AST B1-B3 |

## State mutations and fallbacks

- Test-only local request mutation. No runtime, broker, journal, or LIVE state mutation.
- Clearing a public scalar alone no longer represents deletion of saved state; the private authority must also be absent.

## Safety conclusion

- Safe edit boundary: fixture setup around saved-stop state only.
- High-risk impact: yes; verifies stop non-retreat and exit-authority separation.
