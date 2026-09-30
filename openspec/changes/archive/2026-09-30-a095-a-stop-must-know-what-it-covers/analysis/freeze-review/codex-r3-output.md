**REJECT.** One contradictory safety contract and three freeze-blocking inconsistencies remain. Accepted open questions are not findings by themselves.

Read-only review; no files changed, tests run, network calls, or engine execution. Historical account measurements were checked for correct attribution, not remeasured.

References below: `proposal.md`, `design.md`, `tasks.md`, `issues.md`, and delta paths are relative to `openspec/changes/a095-a-stop-must-know-what-it-covers/`. `canonical/*` means `openspec/specs/*/spec.md`.

**Round-2 dispositions**

| Finding | Status | Evidence and reason |
|---|---|---|
| P0-1 | RESOLVED | False R2-B2 premise and SHALL removed; caller B7/B8 correctly distinguished. Engine-opened baseline assigned to Q3. `design.md:123–136`; `internal/app/engine/adoption.go:108–112`. |
| P0-2 | MOVED — Q4 | Kind, grade, key and reminder-window reconciliation explicitly deferred; replay retained. `design.md:138–147`; `tasks.md:72–76`. |
| P0-3 | PARTIAL | False single-writer SHALL removed; four-function inventory correct, but monotonicity qualifications remain inaccurate—N5 below. `issues.md:12–25`. |
| P0-4 | MOVED — a092 Q7 | Direct unmanaged-report promotion removed; existing critical-call contention correctly transferred, not eliminated in current production code. `design.md:81–102`. |
| P1-1 | PARTIAL | Alerts-off exclusion intended, but unconditional critical SHALL contradicts it—N1. `specs/engine-safety/spec.md:16–27`. |
| P1-2 | PARTIAL | Default off/unlisted and exclude remain normal; include-path compatibility remains Q2(b), subject also to canonical equal-alert rules. `design.md:45–52`; `canonical/exit-policy:81–89`. |
| P1-3 | RESOLVED | Four sites distinguished; fold adapter unreachable in production; two live unmanaged keys must differ. `design.md:104–119`; `reconcileloop.go:338`. |
| P1-4 | RESOLVED | Production ingest deliberately clears its alerter; proposed error-propagation change withdrawn. `tasks.md:62–63`; `reconcileloop.go:333–338`. |
| P1-5 | RESOLVED | Independence and a091 boundary corrected; fixed event-count expectation removed. `design.md:205–222`; `tasks.md:125–141`. |
| P1-6 | RESOLVED | Measurements dated; obsolete deployment prediction removed; fresh measurement required before deployment. `tasks.md:112–115`; `issues.md:62–68`. |
| P2-1 | RESOLVED | Replay now starts at adopted quantity **2**; historical correction explicit. `tasks.md:75`; `proposal.md:100–101`. |
| P2-2 | RESOLVED | “All effective stops exactly −3%” claim removed from active proposal/design. `design.md:156–179`. |
| P2-3 | RESOLVED | Conflicting risk-growth narrative removed; R3 deferred. `design.md:149–154`. |
| P2-4 | MOVED — R3 follow-up | Counterexample retained with provenance uncertainty; source investigation assigned to successor. `issues.md:31–44`. |
| P2-5 | RESOLVED | R3 requirements and implementation tasks removed; remaining task records deferral only. `tasks.md:79–82`; `specs/exit-policy/spec.md:3–7`. |

**Voice A dispositions**

| Finding | Status | Evidence and reason |
|---|---|---|
| A-1 | MOVED — a092 Q7 | Unmanaged exit report remains normal; shared critical-call contention belongs to a092. `design.md:81–102`. |
| A-2 | RESOLVED | Increased quantity explicitly described as protected; separate classification deferred to Q4. `design.md:140–147`. |
| A-3 | RESOLVED | R2-B2 withdrawn; engine-opened quantity checking explicitly in scope. `design.md:123–136`. |
| A-4 | PARTIAL | Single-writer premise removed, but “refresh cannot change value” needs qualification—N5. `issues.md:12–25`. |
| A-5 | MOVED — R3 follow-up | R3 SHALL removed; average-price provenance investigation retained. `issues.md:31–44`. |
| B-1 | MOVED — Q4 | Critical dedupe issue recognized; kind/key decision and replay remain implementation work. `design.md:144–147`; `tasks.md:72–76`. |
| B-2 | RESOLVED | Live unmanaged emit-site key separation required; production fold adapter excluded. `design.md:104–119`. |
| B-3 | PARTIAL | Blanket promotion removed and transition guards retained; alerts-off contradiction remains—N1. `tasks.md:47–61`. |
| B-4 | RESOLVED | Process-lifetime repetition prohibition and misleading outbox-reminder promise removed; retained latch now explicit. `analysis/function-logic/internal-app-engine--reconciledriver.alertunmanaged/branch-test-map.md:9`. |
| B-5 | PARTIAL | Stale/caller bundles repaired; claimed safety conclusions still exceed represented comparisons—N5/N6. `issues.md:14–25`; `tasks.md:86`. |
| B-6 | MOVED — Q3 / R3 follow-up | Undefined quantity baseline is an implementation stop condition; unbounded total-risk-report SHALL removed. `proposal.md:174–175,200–201`; `specs/exit-policy/spec.md:3–7`. |
| B-7 | MOVED — a092 Q7 | All holders of exit-waited locks are covered by a092’s revised requirement, including future a091 contention. `a092…/specs/engine-safety/spec.md:48–50`. |

**Voice B dispositions**

| Finding | Status | Evidence and reason |
|---|---|---|
| B-P0-1 | PARTIAL | Default/exclude promotion withdrawn; alerts-off contradiction survives—N1. `specs/engine-safety/spec.md:16–27`. |
| B-P0-2 | RESOLVED | “No adoption record = unprotected” SHALL/scenario removed; caller evidence corrected. `specs/engine-safety/spec.md:9–10`; `design.md:123–136`. |
| B-P0-3 | RESOLVED | Increased quantity no longer equated with missing protection; Q4 owns classification. `design.md:140–147`. |
| B-P1-4 | MOVED — Q4 | Reminder SHALL NOT and per-key acknowledgement cost explicitly recognized. `proposal.md:176–178`; `tasks.md:72–76`. |
| B-P1-5 | RESOLVED | Two-site key separation now normative; normal exit reporting also prevents outbox preemption. `specs/engine-safety/spec.md:29–33,63–65`. |
| B-P1-6 | PARTIAL | Exit transient report stays normal; reconcile deferrals remain Q2(c), but pending failure reports after successful adoption remain unspecified—N3. `design.md:35–52`. |
| B-P1-7 | RESOLVED | Production nil-alerter wiring recognized and regression task added. `design.md:119`; `tasks.md:62–63`. |
| B-P1-8 | MOVED — a092 Q7 | Direct exit promotion removed; residual shared-lock delay explicitly transferred. `design.md:92–102`. |
| B-P1-9 | RESOLVED | a091’s operator-choice boundary cited; hardcoded count removed. `design.md:54–56,222`; `tasks.md:42`. |
| B-P1-10 | RESOLVED | Refreshed bundles and coordinates match the supplied pinned source; expanded caller/delivery maps present. `tasks.md:28–38`. |
| B-P1-11 | PARTIAL | False normative writer premise removed; remaining inventory qualifications need N5. `specs/exit-policy/spec.md:6–7`; `issues.md:12–25`. |
| B-P1-12 | MOVED — Q3 / R3 follow-up | Unimplementable R3 provenance requirements deleted; quantity schema question explicitly retained as stop condition. `proposal.md:200–201`; `issues.md:40–44`. |

**Round-3 findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| N1 | **P0** | **Alerts-off requirements are mutually unsatisfiable as written.** Adoption enabled + adoption failure + notifications disabled satisfies the unconditional critical/outbox scenario, while another SHALL forbids its gate/mode consequences. Keeping a critical row but bypassing synchronous delivery does not solve this: a124 counts missing transport as failure. | Delta `specs/engine-safety/spec.md:16,26–27,43–57`; `canonical/engine-safety:179,1448–1450`; production nil publisher: `internal/app/engine/notifications.go:83–87`; background handling: `alertdelivery.go:314–323`. Author already recognized the required grading exception in `proposal.md:160–163`. | Put the notifications-enabled prerequisite and operator-choice exclusions directly into the critical SHALL and scenario. Require reports **originating while off** to remain noncritical. Explicitly distinguish previously persisted critical incidents: canonical restart blocking still applies (`canonical/engine-safety:951–968`). Test real notifier/outbox/deliverer wiring, not only an emitter fake. |
| N2 | **P1** | **The delta already answers Q2(c), despite claiming it remains open.** It expressly includes deferred candidates in mandatory critical grading. An implementation answer “deferred is normal” would violate the frozen SHALL. B14 itself proves absence from the adopted set, not a completed failed adoption attempt. | Delta `specs/engine-safety/spec.md:6–8,16–19`; `design.md:35–37,50–52`; `adoption.go:180–185,195–208,210–215`; `judgeHoldings` B14 at `adoption.go:144`. | Exclude undecided deferrals from the frozen grading commitment and corresponding test expectations until Q2(c) is settled. Keep their classification open without publishing an opposite SHALL. Also record canonical include/enabled alert-equivalence constraints on Q2(b), `canonical/exit-policy:83`. |
| N3 | **P1** | **A durable “currently unprotected” report can outlive successful adoption.** This is not limited to quote deferrals: any recoverable adoption failure can create PENDING, then succeed next cycle. Success clears only the memory latch; background delivery can later announce the old present-tense unprotected claim and apply failure handling. No disposition or acceptance test addresses this newly durable lifecycle. | `adoption.go:353–378` versus report body `:410–422`; delivery enumerates PENDING rows and sends stored bodies, `alertdelivery.go:242–261,325–329`; canonical adoption success replaces unmanaged reporting, `canonical/exit-policy:89`. | Define whether this is an immutable historical failure incident or a current protection-status report. For historical delivery, timestamp the failure and prevent misleading current-state wording; specify recovery correlation. Add failure → successful adoption → delayed delivery/restart coverage. Do not silently auto-ack or clear human-owned safety latches. |
| N4 | **P1** | **Q1’s advertised alternatives conflict with frozen edit boundaries.** One permitted answer changes `SeverityOf`’s contract, but the GREEN task prohibits changes to both `SeverityOf` and `Notify`; several bundle conclusions also declare them immutable. | `proposal.md:156–159`; `tasks.md:64–65`; `design.md:60–63`; `analysis/function-logic/internal-obs--notifier.notify/function-logic-map.md:38`. | Make affected-function boundaries conditional on Q1, with renewed maps and review if its contract changes. Alternatively remove that option explicitly. Leaving Q1 open is fine; simultaneously forbidding a listed answer is not. |
| N5 | **P2** | **Four writer functions is correct; their lowering behavior is overstated.** Refresh compares against stored **effective JSON**, not scalar `baseline_price`. It can therefore lower a divergent scalar to the saved protection. Snapshot judgement likewise selects against saved JSON; with no saved snapshot, the selector accepts recomputation without a scalar baseline comparison. These are narrower guarantees than I1 implies. | `issues.md:13–14,24`; `exit_observation_refresh.go:127–142`; `exit_state.go:487–512,648–671`; `exitpolicy/recovery.go:138–150`. Legacy judgement can update scalars without rewriting effective JSON, `exit_state.go:551–557`. | State preconditions: refresh preserves protection **when scalar and effective snapshot agree**; recovery monotonicity is against a valid saved candidate, not unconditionally against the scalar. Record reset as lowering-capable; INSERT initializes rather than lowers an existing row. Add helper/caller evidence before promoting these facts into Q6 SHALLs. This is not a claim that normal production currently reaches divergent-state refresh. |
| N6 | **P2** | **“All lines derive from entry_price” contradicts the included ladder branch.** Rung locks use entry price, but runner protection uses high-water `probe × (1 − trail%)`, and previous baseline participates in the maximum. | `issues.md:25`; `tasks.md:86`; `analysis/function-logic/internal-exitpolicy--evaluateladder/function-logic-map.md:9,43–44`; `internal/exitpolicy/ladder.go:391–403`. | Narrow the claim to entry-relative rung thresholds/locks and frozen R denominator. Preserve high-water runner and previous-baseline behavior in the no-change test. |
| N7 | **P3** | **Transferred questions still appear as active implementation questions.** Transport-dead handling remains in Q2’s design row/task; adoption bundles label deferrals Q2(b), whereas current numbering is Q2(c). | `design.md:52`; `tasks.md:58–59`; `analysis/function-logic/internal-app-engine--reconciledriver.adopt/branch-test-map.md:10,14–15`; `proposal.md:164–173`. | Update these references to Q2(c) and the canonical a124 requirement; leave historical review entries intact. |

The requested production-lock attack **does not establish that the normal unmanaged report itself waits on `n.mu`**: `Notify` takes the normal branch at `internal/obs/notifier.go:134–136`, and `publishBestEffort:161–173` does not acquire that mutex. It still performs synchronous network publication. Other existing critical exit reports can wait behind reconcile delivery (`notifier.go:254–309`); that is the accepted Q7 transfer, not a new dependency recommendation. Tasks 2.1/6.2 prove only the narrower normal-path property.

The cross-citations are accurate: a092 revision 21 requires **all holders** of exit-waited locks to release them during remote transmission (`a092…/specs/engine-safety/spec.md:48–50`; `design.md:451–455`), while its proposal expressly retains reconcile’s own remote wait (`proposal.md:645`). The a124 canonical requirement owns durable failure counting and gate/mode consequences, including absent transport (`canonical/engine-safety:1434–1450`).

R3 removal is consistent: no active total-risk SHALL or implementation task remains. Q3/Q4 work, R2-B2 deletion, key separation, and a092 independence reflect the binding decisions. Decision (2) remains incompletely implemented because of N1.

VERDICT: REJECT — the alerts-off contract contradicts mandatory critical handling, and grading, incident-lifecycle, and edit-boundary inconsistencies must be corrected before freeze.