# Function Logic Map: `StrategyEntrySupervisor.runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| child context/start | one context and shared paired start barrier | `Run` | return on cancellation |
| worker | exactly one constructor-owned KR or US runtime | constructor | impossible state escalates central |
| cycle callback | evaluation-only, bounded by watchdog | descriptor | local latch/restart on ordinary failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | cancelled before start | none | child return | shutdown tests |
| B2 | cancelled while idle/cycle | abandon evaluation-only callback if needed | child return | shutdown/watchdog tests |
| B3 | authority expired | latch only this worker | continue after bounded restart wait | expiry test |
| B4 | cycle succeeds | none | continue same child | concurrent start test |
| B5 | cycle local error/panic/deadline | market-local latch + restart wait | continue child, peer untouched | paired fake-clock restart tests |
| B6 | typed central integrity | signal central | child return/process fail-closed | central integrity test |
| B7 | latch/restart invariant failure | signal central | child return/process fail-closed | saturation/invariant tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluationState` | atomic effective/latch/expiry check | expired becomes local fault | AST B3 |
| `invokeBoundedStrategyCycle` | contain panic and watchdog callback | deadline is abnormal local fault; typed central preserved | AST B4-B6 |
| `latchMarket` | immutable first refusal + saturated restart schedule | only this market mutates | AST B3/B5/B7 |
| `waitMarketRestart` / `clk.Now` / `clk.Sleep` | consume the absolute restart deadline and recalculate its remaining interval immediately before sleep; elapsed deadlines complete without sleeping | negative/zero remainder completes immediately, remainder above the frozen maximum fails central, cancellation ends child, no authority restoration | paired fake-clock handoff-race + saturation tests |

## State mutations and fallbacks

- Market-local failure never returns from the outer loop. Restart only continues the supervised child under the durable/effective OFF latch; it never turns entry back ON.
- Fault delivery and watchdog cleanup may race with the restart wait, but the absolute deadline prevents that latency from moving `RestartNotBefore` forward.
- Central integrity remains the only process fail-closed escape.

## Safety conclusion

- Safe edit boundary: insert a bounded market-local restart wait after the existing latch path, without mutation dependencies.
- High-risk impact: yes — it determines whether peer and safety loops survive entry-worker failure.
