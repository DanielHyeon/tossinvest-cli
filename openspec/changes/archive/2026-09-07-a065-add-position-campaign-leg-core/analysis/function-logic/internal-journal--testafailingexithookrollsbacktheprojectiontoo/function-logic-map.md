# Function Logic Map: `TestAFailingExitHookRollsBackTheProjectionToo`

- Source: `internal/journal/apply_hook_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test fixture | writable journal and synthetic fill | test | fatal on setup mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | setup/apply/assertion branches | test-only rows | test failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| journal test helpers | prove rollback | fatal on error | AST |

## State mutations and fallbacks

- Test-only; no live broker or production toggle.

## Safety conclusion

- Safe edit boundary: atomicity regression test.
- High-risk impact: test evidence for fill rollback.
