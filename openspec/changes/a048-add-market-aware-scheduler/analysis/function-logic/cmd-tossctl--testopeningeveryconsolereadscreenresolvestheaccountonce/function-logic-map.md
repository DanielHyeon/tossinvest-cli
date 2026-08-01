# Function Logic Map: `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scheduler desired document | KR regular market in a temp config directory | test fixture | save failure is fatal |
| counted broker factory | calendar-capable official client and exact account reference | package seam | all read screens together may build it once |
| read-screen set | positions, orders, market schedule | `runConsole` wiring contract | any read error is fatal |
| `console.go` AST | current source | parser | unparseable or unreviewed build/wiring site fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | desired state save fails | none | fatal | self |
| B2-B5 | serial rounds/screens, screen error, or build count mismatch | invokes all three screens twice | fatal/assertion on drift | self |
| B6-B9 | concurrent screens, returned error, or cold build delta mismatch | starts one goroutine per screen | fatal/assertion on drift | self |
| B10-B19 | parse source and enumerate every broker-factory site/allowlist entry | builds site counts | fatal/assertion on unargued or stale site | self |
| B20-B31 | locate `runConsole`, its sole resolver assignment, and each seam's configured shared argument | records holder and seam arguments | missing/malformed calls remain unwired | self |
| B32-B35 | validate one resolver and exact holder passed to every declared seam | none | fatal/assertion on mismatch | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scheduler.DesiredStore.Save` | make market schedule perform an official read | errors are fatal | runtime fixture |
| positions/orders/market-schedule readers | reproduce every live console read screen | screen errors are fatal | serial and concurrent loops |
| `newConsoleBroker` | construct warm and cold shared resolvers | counted factory proves one build per resolver | runtime assertion |
| `parser.ParseFile` / `ast.Inspect` | bind runtime claim to actual `runConsole` source | parse errors are fatal | static wiring assertion |

## State mutations and fallbacks

- Replaces `verifyBrokerFactory` and restores it with cleanup.
- Protects the build counter and calendar fixture so the concurrent half is race-safe.
- The static descriptor names the shared argument index: positions/orders use argument 0; `consoleMarketScheduleSeam(root, reads)` uses argument 1.

## Safety conclusion

- Safe edit boundary: console command integration guard only.
- High-risk impact: medium; duplicate account resolution caused measured 429 failures, but the test performs reads only and no live mutation.
