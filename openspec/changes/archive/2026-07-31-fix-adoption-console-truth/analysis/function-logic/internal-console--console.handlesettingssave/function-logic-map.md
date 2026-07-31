# Function Logic Map: `Console.handleSettingsSave`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Settings seam | non-nil `AdoptionSettings` | injected console option | 501 refusal when absent |
| current config | readable before replacement | seam `Load` | redirect without save |
| `default_stop_percent` | finite 2..20, integer half-percent tick | OpenSpec operator-console delta | specific redirect without save |
| other adoption fields | checkbox and symbol lists | submitted form | preserved in one replacement block |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | settings seam nil | none | refuse 501 | existing unwired test |
| B2 | current config load fails | none | redirect error | existing load-failure tests |
| B3 | percent missing/non-numeric/non-finite | none | redirect; no Save | `TestInvalidStopPercentWritesNothing` |
| B4 | percent outside 2..20 | none | redirect; no Save | same table test |
| B5 | percent not 0.5 step | none | redirect; no Save | same table test |
| B6 | valid half-percent tick | constructs adoption block and calls Save once | success/error redirect | full-grid + happy-path tests |
| B7 | seam Save rejects | seam-defined no-write | redirect error | existing invalid-save test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AdoptionSettings.Load` | refuse replacing unreadable config | no retry | CodeGraph + AST |
| `parseStopPercent` (new leaf) | validate human unit and convert to fraction | deterministic error; no side effect | direct table tests |
| `splitSymbols` | normalize lists | pure | existing tests |
| `AdoptionSettings.Save` | surgical config update + audit in cmd seam | called once only after validation | fake and real-seam tests |

## State mutations and fallbacks

- Mutation is limited to one `Save(next)` call after all validation.
- There is no default/fallback for malformed POST values and no engine restart.

## Safety conclusion

- Safe edit boundary: change only the percentage field/parser; preserve enabled
  and symbol construction, seam call, CSRF route wrapper, and effect notice.
- High-risk impact: yes — protective width setting; config fraction and engine
  consumption remain unchanged.
