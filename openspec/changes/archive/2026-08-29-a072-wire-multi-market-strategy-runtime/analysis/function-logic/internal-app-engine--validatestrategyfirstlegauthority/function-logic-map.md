# Function Logic Map: `validateStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| accepted result | sealed fixed-router KR/US result | prior admission validation | reject revalidation mismatch |
| Guardian issuance | exact scope/q_candidate and major prices | RiskGuardian request | reject before Guardian call |
| campaign metadata | prospective generation and command/activation ids | runtime coordinator | reject incomplete CAS/attempt binding |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | result identity/quantity mismatch | none | authority mismatch error | tamper test |
| B2 | price cannot derive major decimal | none | unit conversion error | malformed unit test |
| B3 | Guardian scope/q_candidate/price/owner mismatch | none | exact-binding error | paired drift table |
| B4 | campaign CAS/generation mismatch | none | campaign error | CAS table |
| B5 | attempt/activation metadata invalid | none | metadata error | metadata table |
| B6 | exact paired authority | none | nil | paired success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateStrategyFirstLegResult` | re-authenticate sealed result | deterministic; no I/O | AST |
| `PriceProvenance.MajorDecimal` | derive Guardian decimal without floats | deterministic; fail closed | pre-edit contract |

## State mutations and fallbacks

- Pure preflight; no Guardian, journal, Gateway, broker, activation, or toggle mutation.
- KR and US use the same code path and test wave.

## Safety conclusion

- Safe edit boundary: validation before first-leg mutation.
- High-risk impact: yes; incorrect US conversion changes notional by 100x.
