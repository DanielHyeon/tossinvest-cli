# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Cobra command/context | configured console invocation | `newConsoleCmd` | nil context becomes background; resolution errors return |
| root/profile options | explicit/default config and credential paths | command flags/config resolvers | fatal setup errors return; optional display seams remain nil with notice |
| shared console broker | lazy unresolved or one resolved account-scoped official client | `newConsoleBroker` | account-scoped read seams report unavailable without widening capabilities |
| instrument-name reader | shared account-resolved client's official `/stocks` surface | `newConsoleInstrumentNames(reads)` | name-only failure leaves journal symbols and frozen values visible |
| console Options literal | fixed capability injection | this function | only enumerated narrow seams cross into `internal/console` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | context or mandatory path/seam resolution fails | no listener/account mutation | use background or return setup error | cmd console setup tests |
| B8-B15 | optional journal/performance/engine-marker setup unavailable | stderr notice; optional capability left nil | continue with explicit unwired/read-only UI | console journal/engine tests |
| B16-B26 | container/local updater capability combinations | bind only verified local update seams | continue with updater disabled or verified | system-update tests |
| B27-B28 | engine lock acquisition | local filesystem lock only | error refuses update | engine lock tests |
| B29-B30 | configured autostart result | optional local process orchestration and notice | continue with result text | autostart tests |
| B31-B36 | position-policy descriptor/dial state | optional loopback RPC reader/commander | continue read-only on absence/error | position-policy console tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| path/config resolvers | bind this profile's durable state | mandatory verification paths return error; optional journal/marker paths degrade visibly | AST B2-B15 |
| `newConsoleBroker` | share one lazy account/client resolution across read seams | no network request until a screen reads | source comment + shared resolver tests |
| `newConsoleInstrumentNames` | expose optional stock metadata through the same account-resolved OAuth manager as other reads | request context bounds cold construction and reads | A061 cold resolver, shared-client, and static boundary tests |
| seam constructors | narrow each console capability | returned interfaces expose only enumerated methods | static capability closure tests |
| `console.ListenAndServe` | build and serve the authenticated console | returns listener/server error; context owns shutdown | AST tail + console command tests |

## State mutations and fallbacks

- Setup writes only operator diagnostics and constructs local capability values.
- Optional dependencies remain nil/unavailable; none are replaced by a wider broker.
- The change adds one read-only `InstrumentNames` field beside `Holdings`, backed by the exact shared official client so optional labels create no competing OAuth manager.

## Safety conclusion

- Safe edit boundary: one Options-literal binding to a new narrow read-only metadata adapter; all existing branches remain byte-for-byte unchanged.
- High-risk impact: yes for capability wiring, mitigated by static method-set closure tests, request cancellation tests, and an adapter with only one exported read method.
