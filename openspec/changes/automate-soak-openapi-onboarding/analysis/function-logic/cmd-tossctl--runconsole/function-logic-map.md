# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Cobra context/root/options | validated command inputs | Cobra/root flags | path or remote validation errors return |
| record/path resolution | profile-local paths | shared app/journal resolvers | required records fail; advisory paths disable panels |
| console seams | narrow read/write/process capabilities | cmd assembly | nil disables only corresponding feature |
| LIVE safety | no direct order capability added | `console.Options` static guards | compile/static tests fail |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil command context | installs background context | continues | command test |
| B2-B6 | remote/record/attestation resolution error | none | early error | resolver tests |
| B7-B9 | journal/engine path unavailable | stderr guidance, disables view | continues | console path tests |
| B10-B20 | container/local updater variants | wires or disables updater | continues with fixed guidance | update wiring tests |
| B21-B22 | engine directory/lock available | wires lock closure | lock error returned by seam | update lock tests |
| B23-B24 | engine autostart seam/note | load/start decision and stderr note | continues | autostart tests |
| final | server exits | graceful container engine stop | returns `finishConsole` result | finish tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| path resolvers | keep one profile namespace | required errors fail early | CodeGraph + AST |
| `console.ListenAndServe` | owns HTTP runtime | returns serve error | current HEAD |
| `restartSoak` closure | narrow process action with post-exit token fence | fence/spawn error returned | CodeGraph |
| `finishConsole` | graceful container shutdown | preserves serve error | current tests |

## State mutations and fallbacks

- Assembly itself does not mutate credentials or accounts.
- New leaf seams own credential source detection, read-only official probe,
  isolated validation cache, protected persistence, token invalidation, and
  secret-free audit.
- `runConsole` only wires those seams into `console.Options`; the restart
  closure passes the token invalidator into the post-exit pre-spawn boundary.

## Safety conclusion

- Safe edit boundary: add one credential seam constructor and two option fields;
  do not change engine, trading, gate, or broker seams.
- High-risk impact: yes, credential/authentication path; no order method exposed.
