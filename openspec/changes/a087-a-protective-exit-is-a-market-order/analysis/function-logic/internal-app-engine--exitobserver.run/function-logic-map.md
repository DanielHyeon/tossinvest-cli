# Function Logic Map: `ExitObserver.Run`

- Source: `internal/app/engine/exitloop.go:407-417`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | live or cancelled | runtime | cancelled context returns unchanged |
| observation cycle | any result | `ObserveOnce` | cycle error is reported, never stops protection |
| refresh token | capacity 1, optional board | a087 board | nonblocking post-cycle signal only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | loop | one observe-report-refresh-sleep iteration | none | refresher order test |
| B2 | context cancelled | none | ctx error | existing cancellation test |
| B3 | sleep cancelled/error | none | sleep error | existing sleep test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ObserveOnce` | complete judgement/submission cycle | errors stay in cycle | AST line 412 |
| `reportCycle` | existing diagnostic output | does not alter loop result | AST line 412 |
| `lowerLimitBoard.cycleComplete` | after observe, before sleep token | nonblocking, no broker call | a087 insertion |
| `clk.Sleep` | preserve cycle cadence | B3 returns its error | AST B3 |

## State mutations and fallbacks

- token send may be dropped while a prior token waits; it never waits behind judgement or submission.
- no board/admission is a no-op, preserving existing cycle latency.

## Safety conclusion

- Safe edit boundary: one nonblocking signal strictly after `ObserveOnce`.
- High-risk impact: yes — ordering proves lower-limit reads cannot overlap a protection cycle.
