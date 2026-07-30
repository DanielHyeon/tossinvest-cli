# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| state-changing paths | exact authenticated action set including update install | route contract | missing/extra CSRF classification fails |
| extracted wrappers | session and mutating/CSRF flags | AST extractor | mismatch fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | iterate routes and reject mutating-without-CSRF or read-with-CSRF | none | test failure |
| B5-B6 | require every declared mutating path to be registered | none | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | inspect real route wrappers | static only | AST |
| `t.Errorf` | report exact route mismatch | deterministic | AST |

## State mutations and fallbacks

- `/settings/system-update/install` is classified as state-changing; path/command fields remain absent.

## Safety conclusion

- Safe edit boundary: add the exact install path to the closed action set.
- High-risk impact: yes — CSRF omission would allow local web content to trigger replacement.
