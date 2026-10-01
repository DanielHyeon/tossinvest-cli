Access rule, unchanged from the previous round: do not read, list or write anything under ~/.codex. Do not access any path outside
the review tree named below. If you did access either at any point, your FIRST line must be `ACCESS-VIOLATION: <what>` before the
verdict line.

This is round 3 of the same freeze review, a narrow re-check. The documents have been revised to "4판" (version 4). The new read-only
tree is `/tmp/claude-1000/a091-r3-tree` (a `git archive` of commit 3c9bf7b0). Read ONLY there. Your previous tree is stale.

Read in the new tree: `openspec/changes/a091-a-stop-that-sold-nothing-is-critical/review.md`. Start at the sections 「Manager 판정
(2026-10-01) — Q1 · Q2 · Q3」 and 「4판」, which map each of your round-2 findings to what changed. Then read `design.md` (4판),
`specs/engine-safety/spec.md`, the NEW `specs/exit-policy/spec.md`, `tasks.md` sections 2–6, `issues.md` and `proposal.md`.

Manager decisions, which are inputs and not up for re-litigation:
- Q1: gate the new critical on the loaded `notifications.enabled` flag only, exactly like a095. "Enabled but transport failing or missing"
  is NOT gated; its latch is intended a092 semantics.
- Q2: a zero caused by account holdings being 0 is excluded (normal). The 8/2 ledger was re-read read-only; the result is in design
  「8/2 원장 재독」.
- Q3: keep a single episode key; the row body carries the episode's first cause with its time, and each observation's cause goes to a
  structured log line.

Verify, with file:line evidence in the new tree:
1. For each of your round-2 findings 1–8, is it now CLOSED, PARTIAL or OPEN? Give one line each, with evidence.
2. Is the new holdings-zero test in design D3 ③ exact? The claim is `Bound == FloorBoundHoldings ∧ Quantity == "0"` ⟺ a fresh holdings
   snapshot of 0 (`internal/riskcalc/confirmed_floor.go` ~129-195). Look for any input that breaks either direction.
3. Is the ctx-cancellation exclusion (D3 ④ / D5) decidable from `ctx.Err()` at B2 without misclassifying a real failure?
4. Does the exit-policy MODIFIED delta faithfully copy the canonical block with only the stated addition? Is it consistent with the
   engine-safety delta?
5. Is the D7 cause contract deliverable by base code? Check `outbox.go` recordAlertTx, `notifier.go` logEvent, and
   `record_only.go` withoutFields.
6. Anything NEW in 4판 that is false, or that weakens §0.3, §0.9 or invariant 3.

Output: first line `VERDICT: PASS` or `VERDICT: FAIL`. Then the per-finding table, then any new findings, each with severity
(P0/P1/P2/P3), evidence and minimal fix. Keep it tight.
