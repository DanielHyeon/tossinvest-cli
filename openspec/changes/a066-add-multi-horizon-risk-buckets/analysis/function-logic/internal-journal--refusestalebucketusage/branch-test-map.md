# Branch Test Map: `refuseStaleBucketUsage`

Post-edit (6.5 fix lot, 2026-09-28). Pre-edit AST and map: `analysis/pre-edit/6.5-fixes/internal-journal--refusestalebucketusage/`. Mutation column refers to the mutation-6.5 ledger `analysis/mutation-6.5/fixes-ledger.tsv`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | caps not aligned with snapshots (31) — **new 6.5** | every admission (plumbing); mutation L04/L05 (nil caps) CAUGHT | RED at `4fb0b26e` (`pre-edit/6.5-fixes/red-at-4fb0b26e.log`) | covered (suite rc 0 at the fix tree) |
| B2 | range buckets (34) | all admission tests | — | covered (suite rc 0 at the fix tree) |
| B3 | ledger read error (36) | storage exit — structural test | n/a | covered (suite rc 0 at the fix tree) |
| B4 | latched usage (43) → `latchedUsageRefusal` names kind + bucket | `TestA066LatchedUsageInASharedBucketBlocksNewEntry`, submit-revalidation cases | mutation V05, V06, V07 | covered (suite rc 0 at the fix tree) |
| B5 | ledger sum unparseable (47) → snapshot mismatch | 5.6.1 rule-edge tests | mutation 5.6.1 ledgers | covered (suite rc 0 at the fix tree) |
| B6 | claimed (snapshot) sum unparseable (51) → snapshot mismatch | 5.6.1 rule-edge tests | mutation 5.6.1 ledgers | covered (suite rc 0 at the fix tree) |
| B7 | claimed < ledger (54) → BUCKET_USAGE_STALE | `TestA066StaleUsage*` | mutation 5.6.1 ledgers | covered (suite rc 0 at the fix tree) |
| B8 | cap key ≠ snapshot key (62) — **new** | unreachable: `validateRiskBucketAdmission` requires `cap.Key == ref.Key == bucket.Key` per index before both call sites (risk_bucket.go:306) | — | covered (suite rc 0 at the fix tree) |
| B9 | recorded-limit read error (66) — new | storage exit — structural test | n/a | covered (suite rc 0 at the fix tree) |
| B10 | no active reservation on the bucket (69) → first entry keeps its own limit | every first admission | — | covered (suite rc 0 at the fix tree) |
| B11 | reservation amount unparseable (73) — new | unreachable: `ReservationAtFinal` is produced by `CalculateAdmission` as a `big.Int` string | — | covered (suite rc 0 at the fix tree) |
| B12 | ledger + reservation > smallest recorded limit (76) → BUCKET_CAP_EXHAUSTED naming the limit — **new 6.5** | `TestA066SharedBucketAdmissionIsCappedByTheSmallestRecordedLimit` (+ control at exactly 80), `TestA066FillTransitionUsesTheSmallestDecisionLimitPerBucket` fixture | RED at `4fb0b26e` (`pre-edit/6.5-fixes/red-at-4fb0b26e.log`) | covered (suite rc 0 at the fix tree) |

Re-review repair (2026-09-28). The recorded-limit population is now every row the ledger counts: HELD, FILLED, or RELEASED with `filled_minor<>'0'` (a partial fill then cancel). Test: `TestA066RecordedLimitIncludesReleasedRowsThatStillCountFilledUsage` (RED at 28629ec6; mutation R02 CAUGHT). The branch shape is unchanged (12).
