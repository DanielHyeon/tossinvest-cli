# Function Logic Map: `TestAnInvalidSaveWritesNothing`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| submitted stop width | exact off-grid `7.6%` | new form contract | any seam save fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | submit authenticated off-grid percentage | counting fake remains unchanged | assertion failure if save count is nonzero | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardHarness.post` | exercise the real percentage parser | local in-memory HTTP only | CodeGraph + AST |

## State mutations and fallbacks

- Only the fake seam is observed; no config, engine, broker, or order.

## Safety conclusion

- Safe edit boundary: use the current field name so the intended off-grid branch is reached.
- High-risk impact: yes; protective-width rejection evidence without live mutation.
