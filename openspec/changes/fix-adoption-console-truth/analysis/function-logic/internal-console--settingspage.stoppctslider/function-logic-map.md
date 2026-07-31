# Function Logic Map: `settingsPage.StopPctSlider`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stored fraction | zero, engine-valid fraction, or rejected raw value | loaded adoption block | zero/invalid renders safe 5% default |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fraction is zero or engine-invalid/non-finite | none | render `5` | default/rejected-block tests |
| B2 | fraction is present | none | render deterministic human percentage | non-default/legacy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fractionPercentText` (new leaf) | decimal percentage formatting | pure, no error | direct render tests |

## State mutations and fallbacks

- Display only; no config mutation or fallback write.

## Safety conclusion

- Safe edit boundary: change return unit from fraction to percentage for the
  replacement numeric control; caller/template changes in the same task.
- High-risk impact: yes — protective-width display, but no engine/config side
  effect.
