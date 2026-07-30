# Function Logic Map: `runVerifyRun`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.list` | boolean | Cobra flag | Returns after rendering the catalogue; no credentials, lock, network, record, or mutation |
| command context | non-nil or background fallback | Cobra | Interrupt cancels the runner but preserves recorded evidence and cleanup |
| root/config and record path | resolvable local paths | `engineJournalDir`, `resolveVerifyRecordFor` | Refuses before broker construction or orders |
| existing evidence plus resume/redo | no settled-step replay without explicit resume/redo | `verifylive.LoadEntries`, `StepCount` | Refuses before broker construction |
| broker/account/runner inputs | official broker, confirmed account, valid supervised plan | `verifyBrokerFactory`, `verifylive.New` | Refuses before `runner.Run` |
| engine execution flock | one holder across engine, update, or verification | `enginelock.Acquire` on `engineJournalDir(root)` | Refuses before account/network/order work; kernel releases on crash |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `--list` | Writes catalogue only | nil | existing list test |
| B2 | command context nil | Uses background context and installs interrupt context | continues | existing verification tests |
| B3 | engine/update already owns flock | no record, account call, or order | exclusion refusal | `TestVerifyRunRefusesWhileSystemUpdateOwnsEngineExclusion` |
| B4 | record path/load fails | none beyond local read | error | existing record-path tests |
| B5 | prior steps exist without resume/redo | none | explicit replay refusal | existing resume tests |
| B6 | broker, recorder, or runner construction fails | closes any opened recorder | error | existing construction tests |
| B7 | runner completes | writes supervised evidence and may send only approved mutations | summary/run error | existing end-to-end tests |
| B8 | runner is canceled/deadlined | preserves evidence and releases marker/flock | nil after interrupted note | existing interrupt tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` + `enginelock.Acquire` | share the engine/update crash-safe exclusion | nonblocking; any failure refuses before live account work | source + `enginelock` tests |
| `resolveVerifyRecordFor` / `verifylive.LoadEntries` | bind the selected market to its evidence record | local errors propagate | source + existing verify tests |
| `verifyBrokerFactory` | resolve the official account/broker | failure propagates before runner construction | source + account tests |
| `verifylive.OpenRecorder` / `verifylive.New` | create durable evidence writer and guarded runner | failure propagates; recorder deferred close | source + runner tests |
| `holdVerifyRunLock` | keep soak/discovery advisory marker fresh | marker failure is logged and run continues by existing contract | `runlock` tests |
| `runner.Run` | execute supervised verification plan | per-step human approval and cleanup remain inside runner | verifylive tests |

## State mutations and fallbacks

- The only account mutations remain inside `runner.Run` after its human approval gates.
- The evidence recorder and advisory marker retain their existing cleanup ordering.
- The new execution flock is an additional fail-closed precondition, held for the
  whole live run and released by defer or by the kernel on process death.

## Safety conclusion

- Safe edit boundary: acquire the shared engine/update flock after the read-only
  `--list` return and context setup, but before record/account/broker work; do not
  weaken runner approval, cleanup, or advisory-marker semantics.
- High-risk impact: yes — this is a live-order verification entry path, so RED
  contention tests, race tests, post-edit AST refresh, and independent review are required.
