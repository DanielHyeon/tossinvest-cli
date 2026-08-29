# Function Logic Map: `Journal.CommitRiskBucketAdmission`

- Source: `internal/journal/risk_bucket.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| admission plan | canonical five buckets, immutable snapshots, exact owner/campaign/lane identity | `riskbucket.CalculateAdmission` plus journal rows | refusal or snapshot mismatch before mutation |
| transaction replay | one immutable `transaction_id` preimage | `risk_bucket_final_decisions` | byte-divergent replay fails closed |
| active owner reuse | same prospective generation/lane/campaign and clean replay digest | owner and state snapshot rows | conflict or replay mismatch |
| entry cleanliness | no exact-market active-owner latch, no exact-market scope latch, no applicable account/symbol reconcile | journal rows inside admission transaction, before owner lookup/insert | `ErrRiskBucketEntryBlocked`; no owner or exposure row written |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | risk calculation, validation, digest or transaction start fails | none | typed refusal/error | existing admission validation tests |
| B5-B8 | same transaction is replayed | none except commit | exact idempotent receipt or replay mismatch | existing admission replay tests |
| B9-B11 | existing reservation is absent, wrong account, or not held | none | reservation/snapshot error | existing reservation binding tests |
| B12 | exact scope has any durable latch or applicable reconcile | none | `ErrRiskBucketEntryBlocked` | both late-fill first-admission regressions |
| B13-B19 | active owner matches, conflicts, is newly inserted, races, or query fails | possible owner insert in transaction | reuse/new owner or fail closed | existing owner collision/race tests |
| B20-B28 | reused owner digest/bucket identity validation | none | replay/snapshot error | existing scale-in identity tests |
| B29-B31 | owner sequence or snapshot-set digest/decision insert fails | transaction rollback | error | existing atomic admission tests |
| B32-B39 | per-bucket policy, snapshot and reservation collision/insert checks | inserts remain transactional | mismatch/error | existing five-dimension admission tests |
| B40-B44 | state snapshot/event record or final commit fails | rollback; otherwise durable admission | error or receipt | existing crash/replay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `riskbucket.CalculateAdmission` | deterministic cap decision | refusal writes nothing | AST plus riskbucket unit tests |
| `verifyRiskBucketStateDigest` | reject tampered owner state before reuse | mismatch blocks admission | replay tests |
| `ensureRiskBucketEntryScopeClean` | enforce durable entry gate before both first-owner INSERT and reuse | exact-market latch/reconcile blocks; other-market evidence stays isolated | both late released-owner fill regressions |
| `recordRiskBucketStateTx` | seal committed admission state | shares admission transaction | crash/replay tests |

## State mutations and fallbacks

- All owner, decision, bucket reservation and state snapshot writes share one SQL transaction.
- The new gate is before active-owner lookup, first-owner INSERT, and every scale-in decision/reservation insert.
- There is no fallback that clears a latch or reconcile state.

## Safety conclusion

- Safe edit boundary: add fail-closed journal cleanliness checks immediately after verifying the reused-owner digest.
- High-risk impact: yes — it prevents new exposure after a durable late-fill/reconcile signal without affecting exit or protection paths.
