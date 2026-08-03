# Function Logic Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| run context/approval | one supervised console run and click confirmer | `Console.startRun` | cancellation settles runner; no implicit approval |
| market/record | normalized KR or US profile record | verify resolver | any path/record error returns before broker construction |
| exclusions | engine flock, profile intent marker, then rate-budget lease | enginelock + runlock + ratebudget | failure refuses the run before official API access |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | engine execution exclusion fails | none | return error | existing console verification exclusion tests |
| B2 | record resolution fails | none | return error | existing console verify path tests |
| B3 | prior record load fails | none | return error | existing corrupt-record tests |
| B4 | profile intent marker path fails | execution lock released | return error | profile path tests |
| B5 | rate-budget lease fails/cancels | marker visible; no broker construction | return error | A061 occupied-budget entrypoint test |
| B6 | broker/account resolution fails | held exclusions released | return error | existing console broker failure tests |
| B7 | recorder open fails | held exclusions released | return error | existing verify record tests |
| B8 | runner construction fails | held exclusions/recorder released | return error | existing verify options tests |
| B9 | runner cancellation/deadline | preserve entries and normalize terminal cancellation | nil run error | existing console cancellation tests |
| tail | runner completes | hold marker and both leases released by defer | summary/entries/error | console verification suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `acquireVerifyExecutionLock` | exclude engine/update/other verification | immediate refusal | existing exclusion tests |
| `holdVerifyRateBudgetIntent` | publish priority before optional metadata can reacquire | unique advisory owner under execution flock | A061 contention tests |
| `acquireVerifyRateBudget` | exclude optional metadata across processes | waits with run context; always before broker | A061 budget test |
| `verifyBrokerFactory` / `verifylive.New` | build supervised live runner | existing approval and safety rails unchanged | console verify tests |
| `holdVerifyRunLock` | ask soak and display reads to yield | advisory; lease remains authoritative | runlock tests |

## State mutations and fallbacks

- New state is the profile intent marker plus lease, both acquired before broker construction and released after runner cleanup.
- No approval, order plan, confirmation, or cleanup behavior changes.

## Safety conclusion

- Safe edit boundary: add cross-process admission ahead of existing live runner construction.
- High-risk impact: yes; the full console verification suite and dedicated budget tests are required.
