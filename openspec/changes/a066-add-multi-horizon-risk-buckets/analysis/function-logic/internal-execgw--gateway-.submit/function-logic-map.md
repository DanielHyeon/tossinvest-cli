# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan/ref | normalized mutation plus durable Guardian reference | public Place/Cancel/Amend adapters and journal | mismatch/refusal settles NOT_DISPATCHED |
| exposure class | derived from mutation shape, never caller label | `mutationPlan.raisesExposure` | entry-only gates; exits bypass |
| final broker boundary | fresh decision/protection/reservation evidence | journal + sealed readiness | any mismatch returns not-sent before `plan.call` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B16 | load/prepare/symbol/decision/protection/entry/preflight/reservation checks | journal attempt may be RECORDED then NOT_DISPATCHED | existing stable refusal | existing gateway suites |
| B17-B20 | fresh decision/key/protection checks in dispatch closure | no broker bytes before all pass | not-sent refusal | existing tamper/protection tests |
| new | fresh q_final/owner/all-bucket revalidation fails | no broker bytes | not-sent refusal | new last-moment tamper broker-spy test |
| B21-B24 | dispatch/nonce/settlement results | dispatch tracking and attempt settlement | stable outcome/error | existing replay/retry suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkReservation` | initial post-decision and final in-closure exposure authority | read-only, fail closed | current AST plus q_final last-moment race test |
| `DispatchVerified` | durable dispatch transition and broker closure | exactly-once/no retry | current gateway contract |
| `plan.call` | sole broker mutation | reached only after final revalidation | broker-spy tests |

## State mutations and fallbacks

- Existing journal Prepare/settle mutations and broker call ordering are preserved.
- Final check is read-only and runs after `checkProtection`; only send-tracker construction remains before `plan.call`, and risk-reducing paths bypass it inside `checkReservation`.

## Safety conclusion

- Safe edit boundary: insert final `checkReservation` after fresh decision/protection verification and before send tracking/broker call.
- High-risk impact: yes — sole broker boundary. The edit only adds a conservative refusal.
