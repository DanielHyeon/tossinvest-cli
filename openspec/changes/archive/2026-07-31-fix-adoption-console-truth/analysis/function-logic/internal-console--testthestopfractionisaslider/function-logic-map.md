# Function Logic Map: `TestTheStopFractionIsASlider`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stored fraction | unset, 0.075, or legacy 0.076 | fake settings block | missing value/label/warning or inline script fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | render each representative value and inspect adoption markup | read-only page render | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardHarness.page` | render the real settings template | local in-memory HTTP only | CodeGraph + AST |

## State mutations and fallbacks

- No settings save occurs; the test only renders and inspects HTML.

## Safety conclusion

- Safe edit boundary: replace the obsolete slider contract with the CSP-compatible numeric percentage contract.
- High-risk impact: yes; protective-width display evidence, no live mutation.
