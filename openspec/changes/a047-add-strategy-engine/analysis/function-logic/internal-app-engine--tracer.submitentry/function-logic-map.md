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
| B1 | exact AST `if` at source line 409: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 430: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 435: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 439: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 451: `if out.State != journal.StateConfirmed {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 453: `if detail == "" && err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

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
