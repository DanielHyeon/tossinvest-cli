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
| B1 | `req.Collect == nil` | none | missing-collector error | `TestRiskGuardianIssueEntryRejectsMissingCollectorBeforeIssuingAuthority` |
| B2 | `g.scopedIntent(req.Intent)` returns error | none | scoped-intent error | cross-account/invalid intent tests |
| B3 | scoped intent side is not BUY | none | entry-only side error | entry/reduction separation tests |
| B4 | `evaluateChain(in)` refuses | refusal observation and optional mode escalation | joined chain/escalation error | chain refusal and escalation tests |
| B5 | `risk.EntryExposureValue(in)` refuses | refusal observation | chain refusal | invalid exposure tests |
| B6 | `req.Collect` returns error | transaction never starts for that attempt | collector error | recollection tests |
| B7 | `exposureUsage` rejects snapshot values/currency | transaction never starts for that attempt | usage error | mixed-currency tests |
| B8 | `req.strategyPlan == nil` | ordinary `RecordDecisionAndReserveWithRecollection` | ordinary issue result/error | existing Guardian issuance suite |
| B9 | `req.strategyPlan != nil` | strategy `RecordStrategyDecisionAndReserveWithRecollection` | strategy issue result/error | activation-gated direct coverage |
| B10 | strategy callback's `collectIssue` fails | no strategy transaction starts | callback error | activation-gated direct coverage |
| B11 | strategy recollection succeeds | copies exact issue and receipt | continue to shared error/success handling | activation-gated direct coverage |
| B12 | selected recollection path returns error | refusal observation only when typed issuance-stage refusal | stable refusal error | ordinary refusal suite; strategy half activation-gated |
| B13 | refusal is typed `*IssueRefusal` at issuance stage | copies stable reason and observes row | return refusal | issuance reason/observation tests |
| Scenario | selected recollection path succeeds | post-commit issued observation; no rollback coupling | immutable decision reference, reservations and optional strategy receipt | ordinary Guardian success; strategy direct success activation-gated |

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
