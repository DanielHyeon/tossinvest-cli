# Function Logic Map: `scanAuditEventUnchecked`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| row | exact audit column order including digest | audit SELECTs | scan error propagated |
| event structure | positive IDs/version, non-blank audit/candidate/key/actor, canonical time, known reason | audit lifecycle schema | corrupt event error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | SQL scan fails | local event only | error | migration/read error coverage |
| B2 | timestamp/identity/version/reason invalid | local event only | corrupt audit event | corrupt legacy/current audit tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| scanner `Scan`, `parseStoredTime` | load exact row and enforce canonical time | one call; strict/no fallback | audit corruption tests |

## State mutations and fallbacks

- Builds a local event and returns persisted digest separately. “Unchecked” skips only digest comparison for controlled legacy migration.

## Safety conclusion

- Safe edit boundary: structural audit decoding.
- High-risk impact: yes; malformed rows must not be re-digested or displayed.
