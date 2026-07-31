# Function Logic Map: `restartSoak`

- Source: `cmd/tossctl/soakproc.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| record path | resolved soak record path | console CLI wiring | derived log path only |
| running soak PIDs | zero or more positive PIDs | `soakFindProcesses` | lookup/signal/wait error prevents spawn |
| pre-spawn fence | nil or token-cache invalidator | Open API console seam | any error after old exit prevents spawn |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | executable path resolution fails | none | error | restart failure test |
| B2 | process lookup fails | none | error | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` |
| B3 | iterate located PIDs | signal eligible old processes | continue | restart tests |
| B4 | PID is this console | skip signal | continue | `TestTheRestartNeverSignalsThisProcess` |
| B5 | old soak signal fails | partial signals possible, no spawn | error | process failure seam |
| B6 | iterate signalled PIDs | wait for clean exit | continue | restart tests |
| B7 | old soak does not exit | no spawn | error | `TestASurveyThatWillNotStopBlocksTheRestart` |
| B8 | pre-spawn fence configured | invalidate token cache | continue or error | token fence tests |
| B9 | pre-spawn fence fails | no spawn | error | `TestTokenGenerationFenceFailureBlocksSoakSpawn` |
| B10 | detached spawn fails | no successful child | error | spawn failure seam |
| B11 | choose truthful result by stopped count | none | success note | restart tests |
| B12 | no old soak | spawn one | created note | `TestRestartingWithNothingRunningJustStartsOne` |
| B13 | one old soak | stop then spawn one | PID note | `TestRestartingTheSoakInterruptsItThenStartsItAgain` |
| B14 | multiple old soaks | stop all then spawn one | PID-list note | multi-process seam |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `soakFindProcesses` | identify old surveys | failure prevents signal/spawn | CodeGraph + AST |
| `soakSignalProcess` / `waitForExit` | end old generation before cache fence | bounded wait; failure prevents spawn | CodeGraph + AST |
| `prepareSpawn` | invalidate shared token cache after old generation exits | fixed secret-free error, no spawn | CodeGraph + AST |
| `soakSpawnDetached` | launch exactly one new read-only survey | error returned truthfully | CodeGraph + AST |

## State mutations and fallbacks

- Old soak processes are signalled and fully awaited before the token-generation
  fence runs.
- The detached child is spawned only after the fence succeeds.

## Safety conclusion

- Safe edit boundary: insert the constructor-bound pre-spawn callback between
  completed old-process waits and the existing detached spawn.
- High-risk impact: yes — this is credential-generation/process lifecycle
  ordering, although the soak remains structurally read-only.
