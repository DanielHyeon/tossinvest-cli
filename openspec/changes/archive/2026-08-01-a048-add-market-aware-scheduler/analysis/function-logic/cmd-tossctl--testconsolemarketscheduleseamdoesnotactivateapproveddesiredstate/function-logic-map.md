# Function Logic Map: `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState`

- Source: `cmd/tossctl/marketschedule_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| approved desired state | enabled/autostart KR regular with versions | test fixture | production activation remains absent |
| typed calendar | authoritative KR response | synchronized stub | missing provenance or wrong request fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | desired-state save fails | none | fatal | self |
| B2 | schedule read fails | none | fatal | self |
| B3 | desired approval leaks into effective activation | none | fatal | self |
| B4 | authoritative calendar provenance is incomplete | none | fatal | self |
| B5 | adapter requests wrong country/date | none | fatal | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scheduler.DesiredStore.Save` | persist approved desired input | errors fail test | store fixture |
| `consoleMarketScheduleReader.Read` | produce dormant effective state and calendar provenance | errors fail test | production seam |
| `stubScheduleCalendar.observation` | read request evidence without a data race | locked snapshot | fixture evidence |

## State mutations and fallbacks

- The desired document is deliberately approved while no activation manifest verifier exists; the expected effective state remains disabled.

## Safety conclusion

- Safe edit boundary: scheduler console provenance test.
- High-risk impact: high safety assertion; desired approval must never be presented as effective runtime activation.
