# Function Logic Map: `GuardianAdapter.IssueAndPlan`

- Source: `internal/strategydispatch/adapters.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| concrete adapters | Guardian, journal, account reader, collector non-nil | dormant approved wiring | reject unconfigured |
| manifest binding | account/policy/limits/settings exact | Guardian + activation | reject mismatch |
| atomic receipt | attempt/account/decision/risk/client/quantity/revision/state exact | committed journal result | reject mismatch, never post-read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | adapter/binding/account read invalid | no issuance/gateway | error | adapter tests |
| B2 | Guardian issuance fails | atomic rollback | error | Guardian/journal tests |
| B3 | receipt binding incomplete/divergent | plan remains recoverable by deterministic id/startup enum | error | binding tests |
| Success | exact receipt | build immutable plan/receipt from committed result | return | integration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RiskGuardian.IssueStrategyEntry` | sole strategy authority issuer | exact sizing + atomic commit | AST |
| committed receipt | avoid fallible post-commit lookup | canonical client id/quantity included | journal tests |

## State mutations and fallbacks

- Does not call the gateway. A returned plan is already durable and recoverable.

## Safety conclusion

- Safe edit boundary: concrete bridge only; no fake Authorize/Plan split.
- High-risk impact: yes, transfers committed authority to gateway plan.
