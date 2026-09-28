# Function Logic Map: `refuseStaleBucketUsage`

- Source: `internal/journal/risk_bucket_usage.go`
- AST evidence: `ast.json` (**post-edit**; the pre-edit copy is in `analysis/pre-edit/6.5-fixes/`; pre-edit was extracted 2026-09-28 at HEAD `4fb0b26e`, 29–57, 6 branches; copy in `analysis/pre-edit/6.5-fixes/`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| buckets | the admission's five snapshots | caller | — |
| ledger usage per bucket | `riskbucket.ReadJournalBucketUsage` (the one computation) | journal rows in the admission tx | read error → snapshot mismatch; latched → `ErrRiskBucketEntryBlocked`; claimed < ledger → `BUCKET_USAGE_STALE` |

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | range buckets (30) | — | — | all admission tests |
| B2 | ledger read error (32) | none | `ErrRiskBucketSnapshotMismatch` wrap | storage exit — structural test |
| B3 | `usage.Latched` (39) | none | `ErrRiskBucketEntryBlocked` — the message does **not** name which latch | `TestA066LatchedUsageInASharedBucketBlocksNewEntry` |
| B4/B5 | claimed or ledger sum unparseable (43/47) | none | snapshot mismatch | 5.6.1 rule-edge tests |
| B6 | claimed < ledger (50) | none | `BUCKET_USAGE_STALE` refusal | `TestA066StaleUsage*` |

## Calls, callers and the planned edit

Callers: `CommitRiskBucketAdmission` (risk_bucket.go:199) and `commitFreshRiskBucketAdmissionTx` (risk_bucket_issuance.go:453).

Planned 6.5 edits:
- (finding 1) B3's message comes from the shared rule function, which names the latch kind(s) and the bucket.
- (finding 3a) A new sibling check runs right after this function at both call sites: the admission's reservation plus ledger usage must fit under the smallest limit recorded on that bucket's active reservations.

Both edits are conservative: entry refusal only.

## Safety conclusion

- Safe edit boundary: the named branches only; entry refusal / fill latch in the conservative direction.
- High-risk impact: yes (admission / submit revalidation / fill path).

## Post-edit (6.5 fix lot)

Post-edit (6.5): gains a `caps` parameter (B1 alignment), the named latch rule (B4 → `latchedUsageRefusal`), and the smallest-recorded-limit cap (B8–B12, via `smallestRecordedBucketLimit`). Pre→post ID map: pre B1→B2, B2→B3, B3→B4, B4→B5, B5→B6, B6→B7. The rest are new.
