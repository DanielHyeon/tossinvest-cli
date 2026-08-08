# Function Logic Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root/config | resolvable engine journal directory | `engineJournalDir` | refuses before account/network/order work |
| context | live run context | `internal/console.runState` | cancellation unwinds runner, recorder, marker, and flock |
| market and redo | normalized market; redo derived by console, not request | `verifylive.NormalizeMarket`, caller | invalid record/plan refuses |
| confirmer | console batch-click approval only | injected `verifylive.BatchConfirmer` | no runner mutation without approval |
| engine execution flock | one holder across engine, update, or verification | `enginelock.Acquire` | refuses while update/engine owns the same flock |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | engine/update already owns flock | no broker/account/order work | exclusion error | `TestConsoleVerifyStarterRefusesWhileSystemUpdateOwnsEngineExclusion` |
| B2 | record resolve/load fails | local reads only | error | existing console verifier tests |
| B3 | broker/recorder/runner construction fails | closes opened recorder | error | existing constructor tests |
| B4 | runner completes | supervised evidence and approved mutations only | summary, entries, error | existing console verification tests |
| B5 | context canceled/deadlined | preserves entries and normalizes cancellation to nil | summary and entries | existing shutdown tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` + `enginelock.Acquire` | share updater/engine crash-safe exclusion across the full run | nonblocking; failure is a pre-account refusal | source + contention regression |
| `resolveVerifyRecordFor` / `verifylive.LoadEntries` | select current market evidence | errors propagate | existing console tests |
| `verifyBrokerFactory` | resolve a fresh official account for the mutation trust context | errors propagate before runner | source + account tests |
| `verifylive.OpenRecorder` / `verifylive.New` | construct evidence writer and supervised runner | recorder closes on every return | existing runner tests |
| `holdVerifyRunLock` | retain existing advisory pause signal for soak | failure remains nonfatal | runlock tests |
| `runner.Run` | execute the click-approved plan | cancellation is normalized only after runner cleanup | verifylive tests |

## State mutations and fallbacks

- The helper still creates a fresh broker per run and never reuses the read-only console client.
- The new flock is held from before account resolution until all runner cleanup
  returns, closing the external verification/update start race.
- Run evidence, advisory marker, approval, and cancellation behavior are unchanged.

## Safety conclusion

- Safe edit boundary: acquire the same engine journal flock before record/account
  work and defer release around the complete existing body.
- High-risk impact: yes — the function reaches live verification orders; contention
  must fail before broker construction and must be reviewed independently.
