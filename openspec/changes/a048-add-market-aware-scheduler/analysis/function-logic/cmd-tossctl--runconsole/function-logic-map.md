# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Cobra context/root/options | parsed command state | `newConsoleCmd` | path/wiring errors return before serve |
| config/data paths | resolved or explicitly unwired | path helpers | verification paths fail; journal/marker degrade with operator note |
| engine/autostart seams | nil or narrow typed seams | command assembly | absent seam stays disabled; refusal is displayed |
| console `Options` | capability-enumerated narrow seams | `console.Options` + static tests | unknown/new authority fails static capability tests |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | context/path/OpenAPI resolution failures | none before listener; returns error | exact resolver error | existing command characterization tests |
| B9-B21 | container/update/updater availability combinations | prints advisory; may wire local update seams | continues with reduced capability | system-update wiring tests |
| B22-B23 | engine directory available | wires advisory lock acquisition | nil when unavailable | engine/update lock tests |
| B24 | engine boot seam available | reads desired autostart and evaluates landed interlock start | note only; no hidden approval | engine autostart tests |
| B25 | `ListenAndServe` returns | container path may stop engine through existing bounded shutdown | `finishConsole` result | finishConsole tests |
| B26 | a048 scheduler desired store path resolves | injects read-only status seam backed by the one shared official client | missing file renders closed defaults; selected market fetches typed calendar on page read | scheduler wiring/default/provenance tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| path resolvers | isolate every local artifact under selected config/data root | mandatory evidence paths return errors; optional panels degrade | CodeGraph + AST |
| `runConfiguredEngineAutostart` | preserve existing separately-approved process lifecycle | startup interlock remains authority | existing autostart tests |
| `console.ListenAndServe` | hand narrow capabilities to local console | blocks until context/server exit | CodeGraph callers/impact |
| `finishConsole` | preserve container engine shutdown contract | original serve error retained | focused tests |

## State mutations and fallbacks

- The function assembles capabilities and starts processes only through existing
  typed seams. a048 adds a read-only scheduler provider. It may call the typed
  official calendar through the already-shared console client only when the page
  reads a selected market; it cannot call candidate, activation or desired save.
- Missing scheduler state is not an error: the page reports OFF/OFF/none/regular.

## Safety conclusion

- Safe edit boundary: one `Options` field plus a local read-only provider built
  from the selected config directory.
- High-risk impact: yes, because this is the console/runtime assembly function;
  regression tests must prove no start/toggle/network side effect was added.
