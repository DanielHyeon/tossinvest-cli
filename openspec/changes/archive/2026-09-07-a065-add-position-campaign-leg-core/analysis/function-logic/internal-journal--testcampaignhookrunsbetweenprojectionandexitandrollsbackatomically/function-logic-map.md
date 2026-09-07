# Function Logic Map: `TestCampaignHookRunsBetweenProjectionAndExitAndRollsBackAtomically`

- Source: `internal/journal/apply_hook_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| hook spies | Project/Campaign/Exit order and forced error | test | fatal on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | setup/invoke/assertion branches | test-only state | test failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| apply hook API | prove ordering and one tx | fatal on mismatch | AST |

## State mutations and fallbacks

- Test-only; zero broker requests and zero runtime toggle mutation.

## Safety conclusion

- Safe edit boundary: tx ordering/rollback regression test.
- High-risk impact: evidence for authoritative fill atomicity.
