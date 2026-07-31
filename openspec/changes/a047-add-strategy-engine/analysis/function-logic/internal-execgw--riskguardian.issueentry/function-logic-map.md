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

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `scopedIntent` | bind account/market/symbol and validate exact risk intent | fail closed | CodeGraph + AST |
| `evaluateChain` / `EntryExposureValue` | Guardian-only risk judgement and exact exposure | evaluated once before issuance | CodeGraph + AST |
| `RecordDecisionAndReserveWithRecollection` | atomic durable authority and aggregate hold | bounded recollection; no orphan decision on failure | CodeGraph + AST |
| `observeEntry` / `escalateFor` | preserve refusal/issued provenance and tighten mode | observation does not delay issuance; escalation never converts refusal to allow | CodeGraph + AST |

## State mutations and fallbacks

- Refused chain: no decision/reservation row; an observation records the refusal.
- Issuance failure: decision and reservation roll back together.
- Success: the issued observation is recorded only after commit.
- a047 must add provenance upstream/downstream without reimplementing this chain.

## Safety conclusion

- Safe edit boundary: preferably unchanged; strategy orchestration supplies a
  canonical intent and separately binds lane/manifest provenance to its durable
  attempt. Any direct edit requires reservation race regression.
- High-risk impact: yes.
