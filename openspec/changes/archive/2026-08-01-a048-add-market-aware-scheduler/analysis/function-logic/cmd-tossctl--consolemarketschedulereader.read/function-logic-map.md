# Function Logic Map: `consoleMarketScheduleReader.Read`

- Source: `cmd/tossctl/marketschedule.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| desired state | strict persisted revisioned state | scheduler desired store | errors return; page renders closed fallback |
| selected market | none, KR, US; combined scope has no single provenance digest | server enum | none skips fetch; unsupported provenance fails closed |
| calendar | shared official typed response adapted at response-completion instant | official client + scheduler adapter | absent/read/parse error exposes no claimed provenance |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | constructor retained path error | none | return error | console path contract |
| B2 | desired load fails | file read only | return error | desired strict tests/UI load error test |
| B3 | market none | no network | closed status with no provenance | closed-default test |
| B4 | calendar reader absent for selected market | none | return error | fail-closed seam contract |
| B5 | scope has no single authoritative market | none | return error | enum/fail-closed contract |
| B6 | injected clock present before request | local clock read | choose deterministic market-local request date | production provenance test |
| B7 | IANA location unavailable | none | return error | clock market tests |
| B8 | official typed read fails | one read attempt | return error | official/console error tests |
| B9 | injected clock present after response | local clock read | stamp response completion | fetched-at completion test |
| B10 | adapter rejects response | no status provenance | return error | scheduler calendar tests |
| success | exact official response adapts | read only | current digest/source/fetched-at status | production provenance test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `DesiredStore.Load` | report desired, never create approval | strict fail-closed read | CodeGraph + AST |
| `TypedMarketCalendar` | fetch current market/date via shared official client | error returned; no fallback source; completion clock sampled afterward | production seam/fetched-at test |
| `AdaptOfficialCalendar` | validate and derive canonical digest | parse errors returned | scheduler calendar tests |

## State mutations and fallbacks

- Only request-time official calendar read occurs; no input/form, state save, activation or runtime start is reachable.

## Safety conclusion

- Safe edit boundary: read-only operator status projection.
- High-risk impact: medium; false provenance could mislead an operator, so failures render unverified defaults.
