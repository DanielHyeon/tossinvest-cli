# Function Logic Map: `settingsPage.StopPctPercent`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| percentage input text | value returned by `StopPctSlider` | settings page block | append display unit only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all inputs | none | deterministic `<value>%` label | render tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StopPctSlider` | obtain human percentage text | pure | CodeGraph + AST |

## State mutations and fallbacks

- Display only; no parse, config mutation, or browser script.

## Safety conclusion

- Safe edit boundary: stop multiplying the now-percentage input by 100; retain
  the explicit percent label.
- High-risk impact: yes — protective-width display only.
