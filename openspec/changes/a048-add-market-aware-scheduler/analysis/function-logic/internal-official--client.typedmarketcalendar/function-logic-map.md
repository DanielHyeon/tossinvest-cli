# Function Logic Map: `Client.TypedMarketCalendar`

- Source: `internal/official/typed_calendar_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| country | trimmed/case-folded KR or US | official endpoint contract | other value returns before request |
| date | empty or exact valid Gregorian `YYYY-MM-DD` | official query contract | malformed/impossible value returns before request |
| payload | typed nullable session fields with RFC3339 times | JSON decoder | malformed time propagates decode error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | country not KR/US | none | validation error | country contract tests |
| B2 | date invalid | none | validation error | exact-date test |
| B3 | valid nonempty date | query only | adds exact `date` | typed calendar fixture |
| B4 | official GET/decode fails | HTTP read only | propagate error | malformed time test |
| success | valid response | official GET | typed response | nullable-session test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateMarketCalendarDate` | shared exact-date gate with legacy read | pure; pre-network | AST + test |
| `Client.get` | authenticated official GET and typed JSON decode | existing error semantics | HTTP fixtures |

## State mutations and fallbacks

- Read-only. Nullable holiday sessions are preserved; malformed time/date never becomes scheduler evidence.

## Safety conclusion

- Safe edit boundary: typed official calendar read.
- High-risk impact: yes, because scheduler consumes this response.
