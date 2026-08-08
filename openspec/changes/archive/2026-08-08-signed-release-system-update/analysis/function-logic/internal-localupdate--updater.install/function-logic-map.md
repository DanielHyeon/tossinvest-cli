# Function Logic Map: `Updater.Install`

- Source: `internal/localupdate/updater.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| reviewed SHA | non-empty lower-case exact candidate digest | POST-hidden reviewed value | `ErrCandidateChanged` |
| current executable | unchanged from process startup | fixed descriptor + startup metadata | `ErrCurrentChanged` |
| candidate/current paths | fixed siblings | updater constructor | never caller-selected |
| commit guard | idle verification/activity | console wiring | refuses before replacement |
| updater mutex | one inspection/stage/install at a time | `Updater.mu` | waits |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unsupported platform/empty or changed SHA | no current mutation | error | existing tests |
| B2 | current changed/guard refuses | no rollback/current mutation | error | existing tests |
| B3 | rollback preparation/publish/sync fails | current retained | error | existing failure tests |
| B4 | current replacement sync fails | rollback restoration attempted | explicit error | existing restore test |
| B5 | success | rollback then current atomically published | result | existing publish-order test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `prepareCandidate` | descriptor-bound candidate copy | validates build/hash | CodeGraph + AST |
| `inspectOpen` | binds current bytes/build | no execution | CodeGraph + AST |
| `restoreRollback` | recovers current after sync failure | reports restoration failure | CodeGraph + AST |

## State mutations and fallbacks

- Add updater mutex before all existing checks; mutation ordering is unchanged.

## Safety conclusion

- Safe edit boundary: serialize the existing algorithm without widening paths or
  changing commit order.
- High-risk impact: yes; executable replacement and rollback path.
