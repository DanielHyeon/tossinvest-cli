# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context/port | context present; loopback port | Cobra | initialization error |
| executable path | canonical self path | `binstamp.SelfPath` | update sections disabled |
| updater | fixed sibling paths | `localupdate.New(self)` | local install/download disabled |
| release downloader | fixed repo/platform/current version/cache | production constructor | signed download disabled with stderr reason |
| engine/verification activity | resolved paths | existing lock helpers | install remains fail closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | record/path resolution fails | warnings or fatal according to existing seam | error/disabled section | existing console wiring tests |
| B2 | local updater construction fails | none | update UI unwired | existing wiring test |
| B3 | signed downloader construction fails | none | signed download UI unwired, local install retained | new assembly test |
| B4 | all wiring succeeds | serves loopback console | `ListenAndServe` result | new real CLI assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `localupdate.New` | binds current/candidate/rollback siblings | no network | CodeGraph + AST |
| `releaseupdate.NewProduction` | binds GitHub/Sigstore/version/cache | no account; network occurs only per POST | CodeGraph + AST |
| `console.ListenAndServe` | owns local HTTP lifecycle | same-port relaunch contract unchanged | CodeGraph + AST |

## State mutations and fallbacks

- Adds one read/download seam. It never constructs an account client and never
  invokes install during assembly.

## Safety conclusion

- Safe edit boundary: construct the fixed release downloader beside the existing
  updater and pass it through `console.Options`.
- High-risk impact: yes; production binary/update wiring requires an assembly
  regression test.
