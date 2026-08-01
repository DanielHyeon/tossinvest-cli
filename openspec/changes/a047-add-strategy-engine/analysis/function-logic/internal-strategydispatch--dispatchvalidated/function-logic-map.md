# Function Logic Map: `dispatchValidated`

- Source: `internal/strategydispatch/dispatch.go`
- Source SHA-256: `43b611261cd8e13ced3490f883c9d1201e1fcf4a128faf3bde39c5d7f27d58eb`
- AST evidence: `ast.json` (`B1` through `B27`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| validated record | complete 60-field scalar view from an already opaque-valid decision | production caller `Dispatch` | stale or post-decision mismatch refuses before official call |
| dependency bundle | non-nil gate, manifest, issuer, official gateway and clock | composition root (deliberately absent in a047) | activation refusal |
| gate/manifest | exact immutable snapshots held through read leases | runtime stores | post-plan change is durably refused as TOCTOU |
| official outcome | actual `execgw.Outcome` plus call error | official gateway adapter | exact terminal state classification; ambiguity is never downgraded to refusal |
| persistence | planned attempt receipt | `StrategyIssuer` journal adapter | any terminal persistence error becomes `ReasonPlan` and retains both causes where applicable |

## Branches and early returns

| Branch | Exact AST condition | Mutation/side effect | Return/error | Direct evidence |
|---|---|---|---|---|
| B1 | any required dependency is nil | none | activation | every missing dependency table |
| B2 | clock is zero or decision is at/past exact expiry | none | stale | zero/exact-expiry rows |
| B3 | initial `ReadGate` fails | none | gate error with cause | injected read error row |
| B4 | initial `checkGate` returns a reason | no issuer/gateway call | exact first refusal | initial gate/precedence table |
| B5 | initial manifest verification fails | no issuer/gateway call | activation with cause | dormant repository row |
| B6 | atomic `IssueAndPlan` fails | issuer called once; gateway zero | Guardian with cause | injected issuer error row |
| B7 | leased revision/binding/decision differs from first snapshot | plan exists; gateway zero | typed TOCTOU | leased binding drift row |
| B8 | leased `checkGate` fails | plan exists; gateway zero | typed TOCTOU | leased kill-switch row |
| B9 | leased activation digest differs from planned digest | plan exists; gateway zero | typed TOCTOU | planned digest mutation row |
| B10 | dispatch clock is zero or reaches exact decision expiry under both leases | plan exists; gateway zero | typed TOCTOU | exact post-plan expiry test |
| B11 | manifest callback errors before the gateway was called | plan exists; gateway zero | typed TOCTOU wrapping manifest error | post-plan revocation and digest-mismatch rows |
| B12 | gate lease completes without error | terminal journal write follows | result selected by B13 | exact confirmed/refused/in-doubt table |
| B13 | classify successful-lease outcome | none itself | B14/B16/B18 arm | successful-lease terminal table |
| B14 | successful-lease disposition is `DISPATCHED` | write dispatched link | nil or B15 | exact confirmed spy rows |
| B15 | dispatched-link persistence fails | persistence attempted once | plan error | injected dispatched persistence failure |
| B16 | successful-lease disposition is `REFUSED` | write refusal | gateway error or B17 | not-dispatched rows |
| B17 | refusal persistence fails in successful-lease path | persistence attempted once | joined plan error | injected refusal persistence failure |
| B18 | successful-lease disposition is `IN_DOUBT` | write in-doubt state | gateway error or B19 | malformed-success rows |
| B19 | in-doubt persistence fails in successful-lease path | persistence attempted once | joined plan error | injected in-doubt persistence failure |
| B20 | lease callback was never entered, or lease error is typed TOCTOU | write TOCTOU refusal | B21/B22 or synthesized TOCTOU | pre-callback error plus typed drift tests |
| B21 | TOCTOU refusal persistence fails | persistence attempted once | joined plan error | injected TOCTOU refusal persistence failure |
| B22 | B20 error is already typed | none | return exact typed TOCTOU | revision/binding/leased-blocker tests |
| B23 | post-call error classifies definitive refusal | write gateway refusal | gateway error or B24 | failed-confirmed official outcome |
| B24 | post-call refusal persistence fails | persistence attempted once | joined plan error | injected post-call refusal persistence failure |
| B25 | post-call error classifies dispatched | none | structurally unreachable | invariant test proves any non-nil post-call error prevents dispatched classification |
| B26 | B25 dispatched persistence fails | none in reachable executions | structurally unreachable with B25 | same invariant; no honest fixture can call this method |
| B27 | post-call ambiguous outcome persistence fails | persistence attempted once | joined plan error; otherwise gateway error | dispatch-started timeout and post-callback-error tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GateReader.ReadGate` / `WithLease` | bind the first snapshot through the call boundary | read failure is pre-plan; lease failure is durably terminal after plan | injected gate fixture and TOCTOU tests |
| `ManifestRepository.Verify` / `WithVerifiedLease` | bind all 32 manifest fields and prevent revoke/call race | writer waits for callback or gateway call stays zero | exhaustive manifest test and writer-barrier lease test |
| `StrategyIssuer.IssueAndPlan` | atomically issue Guardian authority and durable plan | exactly one deterministic attempt; any failure stops before gateway | B6 direct failure test |
| `OfficialGateway.PlaceStrategyEntry` | sole official mutation interface | called only after both leases; no retry here | gateway spy call counts |
| terminal `RecordStrategy*` methods | persist exactly one terminal classification | persistence failure is never hidden | B15/B17/B19/B21/B24/B27 direct error tests |
| `DecisionBinding(payload)` | exact full-record equality | any of 60 fields changed causes activation refusal | exhaustive 60/60 laundering test |

## State mutations and fallbacks

- Planning occurs before any official gateway call. Every reachable post-plan no-call path writes a TOCTOU refusal.
- Exact expiry is checked before planning and again under both leases immediately before the gateway call.
- There is no retry or paper/shadow fallback. Tests use only spies and never reach a real order API or operating toggle.
- B25/B26 are dead under the current classifier contract: this section is entered only when `err != nil`, while `classifyOfficialOutcome` returns `DISPATCHED` only when `callErr == nil`. Removal is a maintenance item before any future activation, not testable behavior.

## Safety conclusion

- Safe edit boundary: dormant orchestration only; a047 provides no runtime installer or manifest minter.
- High-risk impact: yes. The tests directly pin all reachable error/persistence branches and explicitly identify the two unreachable AST branches instead of claiming coverage.
