# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | assembled production engine Context | `engine.NewContext` | subordinate constructors reject missing dependencies |
| `clk` | non-nil production/fake clock | command assembly | passed to all time-sensitive loops and recovery |
| strategy-entry assembly | no input/config/activation capability | `NewDormantStrategyEntrySupervisor` | constructor error aborts runtime assembly |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fill detector construction fails | no runtime starts | exact error | detector branch test |
| B2 | reconcile driver construction fails | no runtime starts | exact error | runtime branch test |
| B3 | exit observer construction fails | no runtime starts | exact error | runtime branch test |
| B4 | restart recovery construction fails | no runtime starts | exact error | runtime branch test |
| B5 | dormant strategy-entry construction fails | no runtime starts | exact error | dormant assembly/static constructor test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFillDetector` | construct official-read fill safety loop | fail assembly | AST B1 |
| `Context.ReconcileDriver` | construct reconciliation safety loop | fail assembly | AST B2 |
| `Context.ExitObserver` | construct exit safety loop | fail assembly | AST B3 |
| `Context.Recovery` | construct pre-loop restart recovery | fail assembly | AST B4 |
| `engine.NewDormantStrategyEntrySupervisor` | construct paired KR/US OFF evaluation loop | takes no config, broker, Gateway, journal writer or activation input | focused dormant assembly test |
| `engine.NewRuntime` | supervise safety loops plus one inert strategy-entry outer loop | existing all-or-nothing semantics unchanged | runtime tests |

## State mutations and fallbacks

- Construction arms existing restart recovery entry latch but invokes no loop or broker transport.
- The strategy-entry assembly is a no-argument helper returning only the dormant paired supervisor.
- No setting, activation manifest, Gateway or mutation dependency can enter the dormant helper.

## Safety conclusion

- Safe edit boundary: add the no-argument dormant constructor call and its one `SupervisedLoop`; do not alter existing loop bodies, recovery ordering or Runtime semantics.
- High-risk impact: yes, because this is production engine assembly; risk is bounded by OFF-only construction and no Cycle/Trigger wiring.
