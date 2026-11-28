# gstack review — Codex 외부 적대 패스 원문(codex-cli, 모델 gpt-6-astra, read-only 샌드박스, 2026-11-23, HEAD 237d6af6a)

Static review at `237d6af6a327999d8b209eb608310fcec79a1e60` against the requested base. No tests executed or files modified.

1. **Acknowledging an old incident can suppress a new scope loss until tomorrow.**  
   [attestation_watch.go:350](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:350) reuses the same outbox key after recovery because it contains only the stage, missing endpoints and date.

   Reproduction within one KST day:
   - 09:00: scope-loss delivery fails, leaving a pending alert.
   - 10:00: a valid replacement becomes healthy.
   - 10:30: the operator acknowledges the original pending alert.
   - 11:00: another replacement loses the same endpoints.

   The watcher resets its in-memory history after recovery, but the durable key still identifies the acknowledged incident. [claimOwed](~/tossos-work/a196-impl/internal/journal/outbox.go:382) suppresses delivery inside the one-hour reminder window. `Notify` returns nil, so [checkOnce](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:292) marks this new occurrence sent. Subsequent hourly checks remain silent even after that window expires; the acknowledged row is also unavailable to pending-alert retries. The new restart-blocking defect receives no notification until the date changes.

   Distinguish recurrence after observed recovery from uninterrupted reissuance. Verify with a real notifier/outbox; the added recovery test uses a recorder that cannot reproduce this suppression.  
   Severity: High  
   Blocks: yes

2. **The prescribed recovery command cannot renew the evidence in the normal production state.**  
   [attestation_watch.go:380](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:380) tells operators to run bare `tossctl verify run`, then await renewal. Two independent prerequisites make that procedure fail:

   - The running engine holds the execution lock throughout its lifetime ([engine.go:209](~/tossos-work/a196-impl/cmd/tossctl/engine.go:209)); verification requires the same exclusive lock ([verify.go:293](~/tossos-work/a196-impl/cmd/tossctl/verify.go:293)). The advertised action therefore fails while the watcher’s engine remains running. Stopping it to proceed interrupts its automatic protection loops, which the instructions do not account for.
   - Even after arranging downtime, an existing verification record makes the bare command fail ([verify.go:324](~/tossos-work/a196-impl/cmd/tossctl/verify.go:324)). Following its `--resume` suggestion does not refresh completed measurements: [Runner.settled](~/tossos-work/a196-impl/internal/verifylive/runner.go:706) skips terminal results regardless of age unless explicitly selected for redo. Subsequent attestation renewal can therefore still omit the expired POST evidence.

   Replace the command-only remedy with an executable maintenance procedure covering protection during downtime, deliberate evidence remeasurement, and confirmation of renewed endpoint coverage before restarting.  
   Severity: Medium  
   Blocks: yes

Recommendation: Fix recurrence identity and recovery instructions before merging because acknowledging an earlier scope-loss alert can suppress a new post-repair scope loss until the next KST day.
VERDICT: FINDINGS