You are an independent ADVERSARIAL senior engineer doing a NARROWED re-verification of revision 8 (8판) of an OpenSpec
change in a Go repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or
delete any file, do NOT run git commands that write, do NOT make network calls, do NOT run the engine or any command that
could place an order, do NOT run Go tests. Reading files and grep/rg are fine.

Repository root: the current directory. It is an export of commit 103f1dedc362fdaf0ec9aade5b5a660dd5bc0b31 with the change
directory overlaid from the author's working tree (identical to that commit). There is no .git directory. The Go files
cited below are byte-identical to the change's pinned base 027163575e78482cdd4d8129f027e928b769ae87.

Change: openspec/changes/a095-a-stop-must-know-what-it-covers/. Do NOT open anything under its analysis/freeze-review/
directory except that this prompt came from there; your verdict must be independent of the other reviewers.

SCOPE — this is deliberately narrow (the manager's instruction). Revision 7 was reviewed in full by two independent voices;
one approved and one rejected with two P1 wording defects. Revision 8 changes only: the two spec deltas' wording, design D1
text, proposal's transfer record, tasks, two new function-logic bundles for internal/config (mergeAdoption,
mergeNotifications) and one line (the error-contract sentence) in 21 generated function maps. Do NOT re-audit the 29
earlier bundles or the whole design. Review ONLY:

1. The revision-8 dispositions recorded in review.md §3.21 (r7b F1, F2, F4, F6, F8; r7a F1/r7b F5 as "record + ordering
   condition"; r7a F2, F5; the refused-notifications-block "refused = off" statement; the error-contract propagation):
   for each, RESOLVED / PARTIAL / NOT RESOLVED against the delta text and the code, one line why.
2. The Q2(a)/(b) interaction specifically: after revision 8, does any SHALL, SHALL NOT or scenario in
   specs/engine-safety/spec.md or specs/exit-policy/spec.md still decide an answer to open questions Q2(a) (grade when the
   adoption settings are refused), Q2(b) (grade of an include-designated attempt failure, including with
   adoption.enabled=false), or Q2(c) (grade of deferred candidates)? Check the "operator-chosen state" definition against
   internal/config/engine.go mergeAdoption and internal/app/engine/adoption.go judgeHoldings (B11/B12) and alertUnmanaged
   (B3–B6); check every scenario's WHEN preconditions against the critical SHALL and the alerts-off SHALL NOT.
3. Mutual satisfiability of the changed sentences: the "refused notifications block is off" SHALL against the
   "notifications-enabled is judged by the loaded config value" SHALL and its SHALL NOT on transport absence
   (internal/config/notifications.go mergeNotifications, internal/app/engine/notifications.go resolveNotificationPublisher),
   and the narrowed retry SHALL ("when the same fact is observed again") against its scenario.
4. The error-contract receipt: spot-check at least five of the 21 changed maps' error-contract line against the function's
   code (results, and whether errors are returned, logged or swallowed); confirm that the seven maps still saying "이 함수는
   그것을 되던진다" belong to functions that do return errors.
5. The ordering condition: is the Q7 first-face amplification stated exactly and consistently in proposal's transfer
   record, design's "남는 경합" section, and tasks 7.0/7.3?

Output (concise, file:line evidence):
- A table for items 1–5.
- New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
  record, P3 = editorial) | finding | evidence | suggested fix. Only findings inside the scope above.
- Final line: `VERDICT: APPROVE` (no P0/P1 left) or `VERDICT: REJECT`, with one sentence why.
