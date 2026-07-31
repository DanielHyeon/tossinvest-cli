# Function Logic Map: `Tracer.submitEntry`

- Source: `internal/app/engine/tracer.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| validated tracer params | one symbol, supported market, exact positive quantity/limit/stop/target | `TracerParams.Validate` in `NewTracer` | wrapped `ErrTracerRefused` |
| issuer/submitter | Guardian entry issuer + official gateway seam | `NewTracer` | construction refusal if absent |
| report pointer | live run report | caller `Tracer.Run` | populated only from gateway outcome |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | market parsing fails | none | tracer refusal | tracer validation tests |
| B2 | Guardian issuance fails | Guardian owns any refusal observation | tracer refusal; gateway zero calls | tracer issuer tests |
| B3 | quantity or limit conversion fails | durable Guardian decision may exist but no broker call | tracer refusal | numeric validation/submit tests |
| B4 | gateway outcome is not confirmed | records outcome IDs; gateway owns durable state | tracer refusal with detail | tracer non-confirmation tests |
| B5 | confirmed | records exact intent/order IDs | nil | `runTracerWithFills` path |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `clock.ParseMarket` | canonical market | fail closed | CodeGraph + AST |
| `EntryIssuer.IssueEntry` | Guardian risk decision + reservation | no bypass or local sizing | CodeGraph + AST |
| `ExitSubmitter.Place` | official gateway mutation | one call; gateway owns in-doubt handling | CodeGraph + AST |
| `floatOf`, `currencyFor`, `accountState` | adapt validated tracer values | conversion error refuses | CodeGraph + AST |

## State mutations and fallbacks

- Guardian issuance occurs before broker submission.
- Client order identity is not supplied; gateway derives it from the decision.
- This path has no retry or alternate submitter and is a reference only; a047
  cannot turn tracer parameters into a strategy configuration surface.

## Safety conclusion

- Safe edit boundary: do not reuse or modify. Mirror its Guardian→gateway
  ordering in a new orchestrator, preceded by immutable manifest revalidation.
- High-risk impact: yes.
