# codex 8라운드(좁은 확인) 출력 원문 (a094 0.5q)

- session `01a0e9d8-0a5d-7110-9039-9b0705e91eb9`, codex-cli 0.154.0, gpt-6-astra, read-only, 트리 `git archive 0b5ea307`
- 2026-09-29 06:07:24~06:10:14 KST, rc 0, tokens 122,745, 401 없음. 프롬프트 `codex-r8-prompt.md`(485933e8, sha256 504649cd…)
- 아래는 `-o` 로 받은 마지막 메시지 그대로다(편집 없음).

---

Read-only review; no edits, tests, network calls, or engine execution.

`C/` = `openspec/changes/a094-a-stop-clears-what-blocks-it/`  
`A/` = `openspec/changes/a092-an-alert-does-not-hold-the-stop/`

| Finding | Status | Reason |
|---|---|---|
| R7-1 | **PARTIAL** | Random streak ID, explicit restart duplication, and clock-repetition tests resolve the design defect; exit-policy’s SHALL still specifies observation-time identity (`C/design.md:20–23`; `C/tasks.md:224–226`; `C/specs/exit-policy/spec.md:39`). |
| R7-2 | **RESOLVED** | a092 entrance is now a hard implementation prerequisite; alternative removed (`C/design.md:87–95`; `C/tasks.md:230`). |
| R7-3 | **PARTIAL** | Audited human recovery is included, but its terminal-evidence and quantity-safety conditions remain insufficient; findings below (`C/design.md:31–43`; `C/tasks.md:239–244`). |
| R7-4 | **RESOLVED** | Accepted Manager interpretation explicitly exempts episode-keyed, zero-window producers, consistent with a092’s named compatibility statement (`C/design.md:48–57`; `A/design.md:759–770`; `C/review.md:982`). |
| R7-5 | **RESOLVED** | Overriding text labels production incidence unmeasured and corrects fixture reachability (`C/design.md:67–70`). |
| R7-6 | **RESOLVED** | Reminder delivery explicitly requires producer re-recording; uninterrupted-delay latch acknowledged (`C/design.md:71–72`; `internal/app/engine/exitloop.go:1682–1685`). |
| R7-7 | **RESOLVED** | Conservative policy replaces the unsupported fault-scope claim; overblocking is documented (`C/design.md:61–63,178–180`). |
| R7-8 | **RESOLVED** | Both cited executable tasks now match the overriding design (`C/tasks.md:167–168,272`). |

New ninth-draft findings:

| ID | Severity | Finding | Evidence | Fix |
|---|---|---|---|---|
| **R8-1** | **P1** | **Evidence fields need terminal-state validation.** All three preconditions can hold while X remains live—the reason Form B exists. Required, nonblank “observed state/time” fields do not reject `OPEN`, cancellation-pending, unknown, or mismatched evidence. As specified, such input can become terminal authority for both clearing and proposal release. | `C/design.md:34–38`; `C/specs/order-execution/spec.md:95`; `C/tasks.md:240,244`. Cancellation acceptance is explicitly insufficient at `C/design.md:266–288`. | Define accepted terminal evidence and reject nonterminal/unknown evidence, invalid observation times, and evidence not bound to X. Add negative tests proving both consumers remain blocked. Audit-before-write must remain mandatory. |
| **R8-2** | **P1** | **Assertion lifetime is not bound to an order incarnation.** Input/storage names broker number X, without specifying durable account/market/day/owner-attempt binding. Even if initial engine attribution is checked correctly, an old assertion can match a later reused number. A non-engine order with no matching engine record is prohibited already; number collisions and reuse are the uncovered cases. | `C/design.md:34–38`; `C/specs/order-execution/spec.md:95`. Existing terminal filtering requires scope and ownership chronology (`internal/journal/fills.go:1877–1889`), and resolves scoped replacement lineage (`:1913–1924`). | Persist the uniquely identified engine order/attempt and canonical scope; bind its CANCEL evidence and lineage. Atomically reject stale/ambiguous targets. Both consumers must match that exact incarnation. Test reused numbers, other accounts/days, ambiguous ownership, and replacement orders. |
| **R8-3** | **P1** | **Truthful terminality does not establish safe remaining quantity.** Example: journal holds 10; original sell fills 4 before cancellation; detector misses those fills. Operator truthfully asserts cancellation. The specified unchanged position plus next-observation stop can submit 10 against 6 remaining. Prohibiting synthetic fills is correct, but no quantity-reconciliation prerequisite replaces their missing evidence. | `C/design.md:37–40`; `C/specs/order-execution/spec.md:199–201`. Normal fill application updates projections atomically (`internal/journal/fills.go:541–547`). Submission uses local held quantity (`internal/app/engine/exitloop.go:1357–1365`); broker floor does not apply without a reconciliation block (`internal/app/engine/exitwiring.go:192–200`). | Keep assertion separate from fills, but require authoritative fill catch-up or a verified safe quantity bound before replacement submission. Test partial/full execution followed by cancellation while detection is unavailable. |
| **R8-4** | **P1** | **Rollback permits loss of post-backup trading history.** After engine startup, restoring the predeployment journal loses attempts, fills, projections, and other safety records—not merely reproducible assertion rows. Re-entering assertions cannot reconstruct broker side effects. | `C/tasks.md:362–366`. Existing restore procedure requires preserving DB/WAL/SHM and explicitly warns about mutations absent from the restored journal (`internal/journal/backup.go:22–35`). | Restrict simple backup rollback to a verified no-post-backup-mutations window. Otherwise preserve current journal files and require controlled recovery/reconciliation before resuming trading. State the complete data-loss boundary in 8.1a. |
| **R8-5** | **P2** | **New streak-ID design was not propagated to the delta requirement.** Implementers still receive contradictory identity contracts. | `C/design.md:20–23` versus `C/specs/exit-policy/spec.md:39`. | Replace observation-time identity with the approved random streak ID and explicit restart semantics. |

Audit-before-write, rejection on audit failure, prohibition of synthesized fills, and additive-only migration requirements are correctly specified (`C/tasks.md:239–244`). Q9-1 approval resolves schema authorization; it does not resolve the safety gaps above.

VERDICT: REJECT — the new assertion can authorize unsafe release/submission, and the rollback contract does not protect post-backup trading history.