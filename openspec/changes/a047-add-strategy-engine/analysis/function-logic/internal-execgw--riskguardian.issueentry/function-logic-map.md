# Function Logic Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| `EntryIssuance.Intent` | scoped BUY `risk.Intent`, exact decimal strings, stop required | risk domain + Guardian account | typed/refusal error, no decision |
| `EntryIssuance.Collect` | non-nil authoritative exposure collector | engine snapshot collector | immediate refusal, no decision |
| policy/limits/clock | immutable Guardian startup snapshot | `NewRiskGuardian` interlock | constructor refusal or chain refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | collector missing | none | error | Guardian refusal tests |
| B2 | scope invalid or side is not BUY | none | error | scoped intent and reduction/entry separation tests |
| B3 | pure risk chain refuses | refusal observation; optional mode escalation | joined chain/escalation error | escalation and observation tests |
| B4 | exposure arithmetic refuses | refusal observation | chain refusal | risk arithmetic tests |
| B5 | collection/usage/reservation issuance fails | at most issuance-refused observation; transaction rolls back decision/reservation | issuance refusal | reservation/recollection tests |
| B6 | atomic decision+reservation commits | post-commit issued observation | immutable decision reference and reservations | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether` |
| B7 | private strategy plan absent | ordinary atomic issuance | normal issued result | existing Guardian suite |
| B8 | private strategy plan present | decision+reservation+strategy lineage+attempt+DISPATCH_START in one transaction | issued result with exact receipt | strategy issuance integration |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `scopedIntent` | bind account/market/symbol and validate exact risk intent | fail closed | CodeGraph + AST |
| `evaluateChain` / `EntryExposureValue` | Guardian-only risk judgement and exact exposure | evaluated once before issuance | CodeGraph + AST |
| `RecordDecisionAndReserveWithRecollection` | atomic durable authority and aggregate hold | bounded recollection; no orphan decision on failure | CodeGraph + AST |
| `RecordStrategyDecisionAndReserveWithRecollection` | atomic strategy authority, hold, lineage and planned attempt | same bounded recollection; no public plan surface | AST + journal tests |
| `observeEntry` / `escalateFor` | preserve refusal/issued provenance and tighten mode | observation does not delay issuance; escalation never converts refusal to allow | CodeGraph + AST |

## State mutations and fallbacks

- Refused chain: no decision/reservation row; an observation records the refusal.
- Issuance failure: decision and reservation roll back together.
- Success: the issued observation is recorded only after commit.
- a047 adds a private strategy-plan branch; ordinary issuance remains unchanged.
- Strategy issuance returns canonical quantity/client-order-id in the same committed receipt, so no post-commit lookup can strand a plan.

## Safety conclusion

- Safe edit boundary: the new branch is reachable only from `IssueStrategyEntry`
  through the unexported `strategyPlan` field and reuses the same risk chain.
- High-risk impact: yes.
