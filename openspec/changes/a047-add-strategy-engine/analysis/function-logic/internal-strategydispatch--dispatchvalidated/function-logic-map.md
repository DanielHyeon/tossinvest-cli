# Function Logic Map: `dispatchValidated`

- Source: `internal/strategydispatch/dispatch.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| validated record | complete 60-field scalar view from already opaque-valid decision | `Dispatch` only in production | any post-decision field mismatch is activation refusal |
| gate/manifest | exact immutable snapshots and read leases | runtime stores | durable refusal after plan |
| outcome | actual `execgw.Outcome` | official gateway adapter | exact terminal map |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | dependencies/time/staleness invalid | none | activation/stale error | invalid tests |
| B2 | initial gate/manifest mismatch | no Guardian/gateway call | stable refusal | gate tests |
| B3 | atomic issue/plan fails | journal handles rollback | Guardian error | issuer tests |
| B4 | gate/manifest/expiry changes after plan | durable REFUSED, gateway call zero | TOCTOU error | TOCTOU/expiry tests |
| B5 | exact CONFIRMED with attempt+broker | durable DISPATCHED links | nil | positive test |
| B6 | empty attempt or definitive failure | durable REFUSED | gateway error | outcome table |
| B7 | sent/unknown evidence | durable IN_DOUBT | gateway error | outcome table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| gate and manifest read leases | close validation/call race | writer waits or call zero | lease tests |
| `StrategyIssuer.IssueAndPlan` | atomic authority before call | one deterministic attempt id | adapter/journal tests |
| official adapter | only `execgw.Gateway.Place` path | preserves full Outcome | compile assertion/tests |
| `DecisionBinding(payload)` exact comparison | binds candidate life/state/times, market/calendar/config/indicator versions, bar source/adjustment, price/drift/HVN/reasons and identity | compile-time field inheritance plus reflection cardinality test | AST + exhaustive laundering test |

## State mutations and fallbacks

- Planning occurs before official call; every post-plan refusal is durable.
- Exact expiry is rechecked under both leases immediately before the gateway call.
- Manifest verification compares all 32 fields; the decision gate compares all 60 current DecisionRecord fields.

## Safety conclusion

- Safe edit boundary: dormant orchestration only; no runtime wiring in a047.
- High-risk impact: yes, official mutation boundary and ambiguity handling.
