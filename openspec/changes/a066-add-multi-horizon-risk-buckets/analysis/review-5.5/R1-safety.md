# R1 — safety semantics (code-security-auditor), reviewed at 90e5170d

Verdict: no P0. No path relaxes without OPERATOR + approval + audit-before-commit on the production path; nothing delays
or refuses stop, exit, reconciliation or fill detection.

| Sev | file:line (90e5170d) | Scenario | Probe / evidence |
|---|---|---|---|
| P2 (latent; fixed in 48b00df5) | risk_bucket_relaxation.go:65 | typed-nil `*audit.Log` passes `auditor == nil`; both releases commit with no audit line | `TestR1TypedNilAuditorCommits`: `err=<nil> releases=1`; latch `owner_overage=0 unknown=1 reservations_latched=0` |
| P2 | risk_relaxation.go:308 → engine :138-147 → transport :87 → CLI :259 | owner state drifts without reseal (fill-failure latch); release fails replay mismatch → HTTP 500 → CLI "outcome unknown" although nothing released; retries fail forever | `TestR1DriftedStateReleaseErrorClass` `replayMismatch=true`; `TestR1DriftRefusalReachesCLIAsOutcomeUnknown` |
| P3 | risk_bucket_entry_loss_lock.go:89-99 | every activation on an open lock writes REAFFIRM; a per-cycle trigger would make approvals always stale and grow the table | not probed; 0 production callers |
| P3 | risk_relaxation.go:189-192, 338-343 + journal.go:174 | release holds the single connection across audit fsync; request-ctx rollback after audit over-reports | not probed |
| P3 | production_snapshot_authority.go:475-501 | after releasing A's latch on an over-limit shared bucket, `Latched=false` but entries are refused by BUCKET_CAP_EXHAUSTED | `TestR1ReleaseWhileSharedBucketOverLimit` |

Checked: only writers of the release tables are the two release functions; `risk_overage_latched=0` writers are the
release and the monotone fill path; lock_seq/event_seq AUTOINCREMENT never reused; BEGIN IMMEDIATE serialises activation
vs release; v34→v35 in one transaction; releasing one owner cannot unlatch another owner's rows (probe); UNKNOWN and
overage_minor kept (probe `unknown=1 overage_minor=28`); lock judged only on EXPOSURE_RAISING (`gateway.go:889-891`).
Repo status at end: modified files belonged to the implementer's concurrent repair (not R1). Copy deleted.
