# Function Logic Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Judges an entry and, if the chain allows, issues the decision and its reservation in one journal transaction. This change added three observation hand-offs and changed nothing else.

Placement is the whole point (design D6): the issued-path hand-off is **after** the commit, so a measurement failure cannot roll the decision back; and it is **non-blocking**, so a slow journal cannot sit in front of the caller's Gateway submission and consume the 60-second decision TTL.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.Collect` | non-nil | caller | returns an error before anything is evaluated |
| `req.Intent` | Side must be BUY; scoped to the Guardian's account | caller | `scopedIntent` error, or a BUY check |
| `req.Account` | the state the chain measures against | caller | the chain refuses with its own code |
| `g.observer` | nil or an `EntryObserver` that does not block | construction | nil records nothing; the issuance is unaffected either way |
| `g.costs` | `Configured()` — enforced at construction | construction | n/a — a Guardian cannot exist without one |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/execgw/riskguardian.go:328) | `req.Collect == nil` | none | error; nothing observed — no verdict was reached | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether` (negative case in riskguardian_test.go) |
| B2 (`if` @ internal/execgw/riskguardian.go:332) | `scopedIntent` error | none | error; nothing observed | riskguardian_test.go scoping cases |
| B3 (`if` @ internal/execgw/riskguardian.go:335) | `intent.Side != SideBuy` | none | error; nothing observed — `IssueReduction` is the exit path | `TestAReductionIsNotObserved` |
| B4 (`if` @ internal/execgw/riskguardian.go:349) | the chain refused (`!verdict.Allowed`) | **observes** `REFUSED_CHAIN` with the step and reason | `chainRefusal` joined with any escalation | `TestARefusalIsRecordedForTheFirstTime` |
| B5 (`if` @ internal/execgw/riskguardian.go:358) | `risk.EntryExposureValue` refused | **observes** `REFUSED_CHAIN` | `chainRefusal` | `TestARefusalIsRecordedForTheFirstTime` (same arm, exposure variant) |
| B6 (`if` @ internal/execgw/riskguardian.go:394) | the recollection closure's `req.Collect` error | none | propagated out of the issuance | riskguardian_test.go recollection cases |
| B7 (`if` @ internal/execgw/riskguardian.go:398) | `exposureUsage` error inside the closure | none | propagated | riskguardian_test.go |
| B8 (`if` @ internal/execgw/riskguardian.go:427) | `RecordDecisionAndReserveWithRecollection` failed | **observes** `ALLOWED_ISSUANCE_REFUSED` with the ledger's stable reason, but only when the refusal is `StageIssuance` | `issuanceRefusal(err)` | `TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome` |
| B9 (`if` @ internal/execgw/riskguardian.go:434) | the refusal is an `*IssueRefusal` at `StageIssuance` | guards the observation so a non-refusal error (a bug or outage, per `IssueRefusalReason`) is not recorded as a verdict | n/a — inner guard | `TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluateChain` | the chain, exactly once | pure; no timeout | riskguardian.go:66 + design D1 |
| `risk.EntryExposureValue` | the exposure this entry adds | pure | AST B5 |
| `risk.MeasureEntry` (via `entryObservation`) | the two ratios and the break-even | **cannot fail** — fills what it can, leaves the rest empty, so no branch here can turn a measurement problem into a refusal | netrr.go |
| `g.costs.Fingerprint()` | identifies the rate set the ratios were computed under | pure | costs/fingerprint.go |
| `g.journal.RecordDecisionAndReserveWithRecollection` | the atomic issuance | bounded recollection (3 attempts / 10s), under the 60s TTL | issuance.go |
| `g.observeEntry` | hands the row to the observer | **must not block** — `AsyncObserver` enqueues on a bounded channel and drops when full | observation.go |

## State mutations and fallbacks

- Writes: the `decisions` and `risk_reservations` rows, inside the atomic transaction. Unchanged by this change.
- Writes: one `entry_decision_observations` row per verdict, **outside** that transaction and on the observer's own goroutine. A failure of it is counted in an independent store and never returned to this function.
- Fallback: none added. A nil observer is a no-op, which is why `TestAGuardianWithNoObserverIssuesAsBefore` can assert the pre-change behaviour exactly.
- Live binding: the observer is injected at construction (`RiskGuardianOptions.Observer`), so production wiring chooses whether measurement happens at all.

## Safety conclusion

- Safe edit boundary: the three hand-offs are statements with no return value and no error, placed at points where the verdict is already decided. There is no control flow from an observation to a return.
- High-risk impact: **yes** (Guardian issuance path). §0.9 not engaged — no ALLOW/REFUSE changed; the 272-test execgw suite passes unmodified.
- Ordering proof: `TestASlowObservationDoesNotDelayTheIssuance` times an issuance against a 2-second observer and asserts the issuance does not lengthen; `TestAFailedObservationDoesNotChangeTheVerdict` asserts a failing write leaves the decision and its reservation on disk.
- Upstream inheritance: none — the Guardian is TossOS-new.
