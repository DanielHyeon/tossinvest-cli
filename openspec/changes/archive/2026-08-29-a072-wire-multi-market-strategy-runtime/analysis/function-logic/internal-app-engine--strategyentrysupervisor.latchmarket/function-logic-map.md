# Function Logic Map: `StrategyEntrySupervisor.latchMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| worker | exact constructor-owned KR or US runtime | `runMarket` | impossible state becomes central error |
| failure/abnormal | non-central local failure; abnormal distinguishes panic/deadline | bounded cycle result | canonical fallback reason/code |
| clock | non-zero UTC observation | injected clock | central integrity error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | latch revision exhausted | no partial handoff | central error | invariant test |
| B2 | first local failure | effective OFF, latch, immutable reason/refusal/latch ID/revision | fault + absolute restart deadline | paired fault tests |
| B3 | reassembled worker carries earlier typed refusal | preserve first refusal while storing this assembly's first reason/latch identity | fault uses immutable typed value | immutability test |
| B4 | restart attempt below max | increment and bounded step deadline | local restart schedule | paired fake-clock test |
| B5 | restart attempt at max/time near max | saturate counter/deadline/backoff | local restart schedule | saturation test |
| B6 | zero observation time/fault channel full | no authority restoration | central error | central invariant coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clk.Now` | trusted observation/restart base | zero fails central | AST B2/B6 |
| restart helper | saturating attempt/backoff/deadline | maximum bounded backoff | new tests |
| fault channel send | hand off immutable market evidence | saturation fails central | existing handoff tests |

## State mutations and fallbacks

- Mutates only the failing market runtime under one lock. Peer market fields are not read or written.
- First typed refusal, human-readable reason and latch identity are write-once; restart counter/deadline may advance but never re-enable entry.
- The returned value is the published absolute `RestartNotBefore`, so fault handoff latency cannot silently extend the schedule.

## Safety conclusion

- Safe edit boundary: market-local in-memory observation fields and bounded restart schedule only.
- High-risk impact: yes — a wrong branch could stop the peer/safety loops or accidentally restore entry.
