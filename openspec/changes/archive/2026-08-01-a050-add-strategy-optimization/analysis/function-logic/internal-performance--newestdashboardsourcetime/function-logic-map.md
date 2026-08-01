# Function Logic Map: `newestDashboardSourceTime`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| persisted timestamp strings | blank or strict RFC3339Nano trade-close/metric-observation timestamps | performance DB rows | blank ignored; malformed non-blank value returns error |
| output | maximum valid timestamp normalized to UTC, or zero if every input blank | persisted values only | never substitutes current time |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate every candidate source timestamp | local maximum only | continue | multi-source maximum test |
| B2 | trimmed value is blank | none | skip | blank optional timestamp test |
| B3 | RFC3339Nano parse fails | none | wrapped invalid persisted timestamp error | corrupt timestamp test |
| B4 | parsed timestamp is after current maximum | update local UTC maximum | continue | ordering/timezone test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | recognizes optional blank persisted values | pure | blank test |
| `time.Parse(time.RFC3339Nano)` | strict persisted-format validation | one parse per non-blank input; any error fails closed | corruption test |
| `UTC`, `After` | deterministic maximum independent of offset | pure | timezone/ordering test |

## State mutations and fallbacks

- Only a local maximum is updated. There is no database write or wall-clock fallback.
- A malformed non-blank timestamp invalidates the evidence row rather than being ignored as stale/zero.

## Safety conclusion

- Safe edit boundary: strict persisted freshness derivation.
- High-risk impact: yes for evidence freshness gating; no trading authority.
