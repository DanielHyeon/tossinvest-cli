# Function Logic Map: `TestSavingTheFormWritesTheBlock`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| submitted stop width | `7.5` human percent | HTML form contract | no save or wrong fraction fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | authenticated CSRF save succeeds | fake settings save count increments | test failure on status/value mismatch | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardHarness.post` | exercise the real handler and conversion | local in-memory HTTP only | CodeGraph + AST |

## State mutations and fallbacks

- Only the counting fake is mutated; no real config, engine, broker, or order.

## Safety conclusion

- Safe edit boundary: change the test input unit and expected stored fraction together.
- High-risk impact: yes; it verifies a protective-width conversion without live side effects.
