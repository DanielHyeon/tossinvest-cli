# Function Logic Map: `Journal.RecordQFinalCampaignFirstLeg`

- Source: `internal/journal/strategy_first_leg_atomic.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal | non-nil open v26 journal | `Open` | explicit error, zero writes |
| prospective token | absent from caller request | journal crypto entropy inside BEGIN IMMEDIATE | caller token or entropy failure aborts before inserts |
| q_final issue | exact exposure-raising RiskIntent, aggregate reservation, five sealed buckets | q_final preparation and `riskbucket.CalculateAdmission` | typed refusal/mismatch, zero writes |
| strategy plan | complete immutable decision/attempt lineage bound to the same RiskIntent | strategy lineage verifier | exact binding error, zero writes |
| campaign/leg | canonical campaign/command/plan IDs, current position generation/version | PositionCampaign CAS rules | generation/claim conflict, full rollback |
| router identity | exact `strategyrouter.RouterID` and `strategyrouter.RouterRelease`; both included in request and binding digests | sealed production router constants | missing, forged or stale-release input fails before transaction; divergent replay fails closed |
| replay | one exact v26 binding or no related identity | `strategy_first_leg_bindings` + q_final replay verifier | divergence/partial authority fails without repair |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | nil journal, caller token, invalid campaign, non-production router ID/release, begin failure | none | invalid/begin error | paired fresh-forgery/input guards |
| B5-B7 | replay lookup failure or fresh token mint/entropy failure | none | lookup/token error | entropy and collision tests |
| B8 | q_final/strategy/campaign preparation fails | none | typed validation/refusal | cross-market/substitution tests |
| B9-B13 | replay q_final absent/divergent, binding mismatch or replay commit failure | read transaction only | replay mismatch/error | exact/divergent/partial replay tests |
| B14-B16 | fresh path finds prior q_final or replay lookup error | none | partial authority mismatch | no-repair tests |
| B17-B25 | aggregate precheck, decision/reserve, q_final, lineage, campaign/leg, binding or commit failure | preceding writes inside same transaction only | error + deferred rollback | late-statement failure matrix |
| success | all exact authorities pass | all families + v26 binding commit | opaque receipt | paired KR/US and same-scope race tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `firstLegReplayTokenTx` / `mintFirstLegToken` | reuse only the stored token on exact retry; mint 256-bit fresh authority otherwise | no caller fallback | AST B5-B7; token tests |
| `prepareQFinalCampaignFirstLeg` | recompute and seal q_final, strategy, campaign and router ID/version cross-family preimage | pure, fail closed | AST B8; substitution/router replay tests |
| `recoverQFinalIssueReplayTx` / `verifyFirstLegReplayTx` | validate complete exact replay without repairing rows | no retry inside transaction | AST B9-B16; replay tests |
| aggregate/q_final insert helpers | commit decision, aggregate, owner and five HELD buckets | BEGIN IMMEDIATE; any error rolls back | AST B17-B21 |
| strategy/campaign/binding helpers | append strategy attempt, campaign/claim, leg 1 and immutable companion | no dispatch/execution lineage | AST B22-B24; rollback tests |
| `tx.Commit` | publish the full authority set atomically | no automatic commit retry | AST B13/B25 |

## State mutations and fallbacks

- Fresh success writes exactly one decision, aggregate reservation, q_final/owner/five buckets,
  strategy decision/attempt, campaign/claim, first leg, command/events and v26 binding, including router ID/version.
- No `strategy_execution_lineage`, dispatch lease, Gateway, broker, activation or toggle row is written.
- Exact replay is read/verify/commit only; missing rows are never reconstructed.
- Every failure uses the deferred transaction rollback. Bounded recollection exists only in the
  outer wrapper and retries snapshot-stale/superseded errors with a fresh caller collection.

## Safety conclusion

- Safe edit boundary: package-private preparation and insert helpers remain under the one public
  journal transaction; do not split them into independently committing APIs.
- High-risk impact: yes — paired KR/US, replay, rollback, race, schema trigger and future-lease guards
  are mandatory before production bridge work.
