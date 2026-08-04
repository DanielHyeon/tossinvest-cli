# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | assembled production engine Context | `engine.NewContext` | subordinate constructors reject missing dependencies |
| `clk` | non-nil production/fake clock | command assembly | passed to all time-sensitive loops and recovery |
| strategy-entry production assembly | one command context, request context and frozen runtime clock; internally loads KR/US desired state, official calendars and signed activation in one paired operation | `Context.NewPairedStrategyEntryProductionAssembly` | each market's invalid schedule authority is refused independently; missing candidate/risk/protection completeness still keeps both workers dormant; constructor error aborts runtime assembly |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fill detector construction fails | no runtime starts | exact error | detector branch test |
| B2 | reconcile driver construction fails | no runtime starts | exact error | runtime branch test |
| B3 | exit observer construction fails | no runtime starts | exact error | runtime branch test |
| B4 | restart recovery construction fails | no runtime starts | exact error | runtime branch test |
| B5 | paired production schedule collection or strategy-entry construction fails | no runtime starts | exact error | paired assembly/static capability test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFillDetector` | construct official-read fill safety loop | fail assembly | AST B1 |
| `Context.ReconcileDriver` | construct reconciliation safety loop | fail assembly | AST B2 |
| `Context.ExitObserver` | construct exit safety loop | fail assembly | AST B3 |
| `Context.Recovery` | construct pre-loop restart recovery | fail assembly | AST B4 |
| `Context.NewPairedStrategyEntryProductionAssembly` | collect KR/US schedule authority at one frozen observation and construct the paired production seam | official calendar reads start concurrently; public result exposes scalar schedule verdicts only; missing candidate/risk/protection completeness keeps both workers OFF and mutation-incapable | focused paired schedule/assembly and static capability tests |
| `engine.NewRuntime` | supervise safety loops plus one inert strategy-entry outer loop | existing all-or-nothing semantics unchanged | runtime tests |

## State mutations and fallbacks

- Construction arms existing restart recovery entry latch but invokes no loop or broker transport.
- The command passes the request context and runtime clock to the engine-owned paired assembly; it does not parse activation files, select one market first or construct a market-specific worker itself.
- The engine reads exact KR/US desired and signed activation artifacts plus official calendars concurrently. It exposes only read-only scalar schedule verdicts back to the command; opaque activation authority remains private.
- No activation manifest writer, operational toggle, raw broker mutator or direct order method can enter the public assembly result.
- Until signed per-market activation plus candidate/risk/protection authority are assembled, both KR and US remain `Effective=false`, `Cycle=nil` even when ordinary automation facts are true.

## Safety conclusion

- Safe edit boundary: replace the command's value-only constructor call with the context-owned paired loader/assembly and its one `SupervisedLoop`; do not alter existing loop bodies, recovery ordering or Runtime semantics.
- High-risk impact: yes, because this is production engine assembly; risk is bounded by one frozen clock, closed exact KR+US cardinality, market-local schedule refusal, OFF-only workers and no Cycle/Gateway wiring at this checkpoint.
