# codex 9라운드(좁은 확인) 출력 원문 (a094 0.5s)

- session `01a0e9ed-c575-7693-8c53-2771dc0e58a9`, codex-cli 0.154.0, gpt-6-astra, read-only, 트리 `git archive e25d9b36`(17534)
- 2026-09-29 06:31:08~06:33:24 KST, rc 0, tokens 114,767, 401 없음. 프롬프트 `codex-r9-prompt.md`(164eb417, sha256 36e8bfcb…)
- 아래는 `-o` 로 받은 마지막 메시지 그대로다(편집 없음).

---

Read-only review. No files changed; no tests, network calls, or engine execution. `C/` denotes `openspec/changes/a094-a-stop-clears-what-blocks-it/`. References use this export’s line numbers.

| Round-8 finding | Status | Reason |
|---|---|---|
| R8-1 | **MOOT** | Assertion authority withdrawn; D−7.2 explicitly retracted, associated requirements/scenarios removed (`C/design.md:30–32,85`; `C/tasks.md:244–245`). |
| R8-2 | **MOOT** | No assertion persists or can match a later order incarnation; additive table approval withdrawn (`C/design.md:30–32`). |
| R8-3 | **MOOT** | Assertion-driven release/submission removed. Terminal evidence now explicitly means “종결 체결 기록뿐” (`C/specs/exit-policy/spec.md:41`). Replacement recovery argument has a separate defect below. |
| R8-4 | **RESOLVED** | Simple restore restricted to verified zero post-backup mutations; otherwise preserve DB/WAL/SHM and reconcile before resumption (`C/tasks.md:364–365`; `internal/journal/backup.go:22–35`). |
| R8-5 | **RESOLVED** | Delta now specifies a streak identifier; executable task retains random-ID and restart semantics (`C/specs/exit-policy/spec.md:39`; `C/tasks.md:226–228`). |

Retraction is consistently propagated through the requested documents, including withdrawn 4.Tb/4.Tc/4.Td/8.1a and new 3.R7d–3.R7f. The review correctly labels draft 10 a revision, not approval (`C/review.md:999–1011`).

**Retraction argument: justified decision, overstated proof.**

- **No existing single-order detector catch-up entry point:** correct. `Refresh` invokes `PollOnce`; its comment rejects topic-specific partial refreshes (`internal/filldetect/hints.go:281–289`). That comment does not establish a universal prohibition on every future, separately specified quantity-recovery mechanism.
- **Full cycle all-or-nothing:** false. Collection/read failures prevent application, but application commits **per snapshot**. A later `Ledger.Apply` error leaves earlier commits intact (`internal/filldetect/detect.go:293–321`; `internal/journal/fills.go:339–343,541–553`). The mutex serializes one detector instance, not CLI and engine processes (`detect.go:172–175,277–280`).
- **X is tracked:** yes, for an unambiguous, correctly scoped confirmed PLACE/AMEND. Unobserved orders and nonterminal snapshots are included; CANCEL confirmation does not remove them (`internal/journal/fills.go:1550–1567,1596–1609,1634–1646`). Ambiguous ownership is guarded, not proof of normal tracking (`fills.go:1727–1737`).
- **Successful cycle settles X:** **not guaranteed**. A successful read can report X still OPEN. More seriously, `Apply.FailClosed` is accepted without a cycle error, followed by `outage.success` (`detect.go:324–327,346–356`). Neither outcome establishes terminality.
- **Could a truthful assertion safely change an outcome?** In a particular state, yes: X actually cancelled with zero executions, accurate local holdings, but its readable broker record still reports OPEN. Cycles succeed without terminal evidence; truthful terminal authority could change release without stale-quantity overselling. The withdrawn design did **not prove those conditions**, so this is no recommendation to restore it. It refutes only “assertions cannot change the outcome in any case.”

**Invariant and human-path assessment.**

Holding under uncertain quantity is a sound conservative action: submission uses local holdings, and the broker floor is conditional on reconciliation blocking (`internal/app/engine/exitloop.go:1357–1365`; `exitwiring.go:192–200`). It does not, by itself, prove stop-loss immediacy or eventual recovery. D−8.2 conflates avoiding overselling with closing the liveness obligation.

The named alerts are conditional:

| Alert | Actual reachability |
|---|---|
| Delay | Repeated eligible clearing failures reach `noteDelay`; threshold and one-shot latch apply (`exitloop.go:1223,1247–1256,1675–1685`). |
| D−2.7 counter | Planned trigger counts repeated `clear=false`; CONFIRMED cancellation is not excluded (`C/design.md:350–351,795–806`). It is not yet an existing implemented counter. |
| Filldetect degradation | Wired correctly, but requires consecutive **cycle failures**, not missing terminal evidence (`cmd/tossctl/engine.go:695–699,760–766`; `runtime.go:420–433`). |

Thus detection can appear healthy indefinitely while the hold persists. A concrete example is repeated readable `CLOSED` records lacking cancellation/full-fill evidence: derivation refuses them, `RecordFill` returns a durable refusal with **nil error**, and the detector resets outage health (`internal/brokerstate/derive.go:567–573`; `internal/journal/fills.go:384–409`; `detect.go:324–347`). Repairing authentication/network access cannot resolve that evidence deficit.

Silence is also possible **for these three named alerts**: clearing alerts need eligible observations, while successful detector cycles prevent degradation. Missing/unusable quotes skip judgment, and an already-pending stop suppresses a new proposal (`exitloop.go:451–462`; `internal/exitpolicy/ladder.go:439–443`). Other diagnostic alerts may exist; this does not establish total system silence. The older conditional wording was more accurate (`C/design.md:251–252`).

New findings are limited to draft 10’s replacement argument:

| ID | Severity | Finding | Evidence | Fix |
|---|---|---|---|---|
| R9-1 | **P1** | “Repair detection → one successful cycle → terminal evidence → release” is not a complete human recovery path. Healthy cycles can indefinitely retain OPEN or refused evidence. | `C/design.md:24–26,44–50,54–59`; `detect.go:324–347`; `fills.go:384–409` | Specify recovery/escalation for healthy-but-nonterminal and refused-evidence states; distinguish transport health from terminal-evidence readiness. Add acceptance cases proving the supported recovery path, preserving quantity safety. |
| R9-2 | **P2** | Whole-cycle atomicity claim is false. | `C/design.md:23`; `C/review.md:1003`; `detect.go:316–321`; `fills.go:541–553` | State collection completeness and per-order transactional application separately; remove the binary success/failure proof. |
| R9-3 | **P2** | “Three alerts fire during the hold” overstates coverage; body-only changes do not establish reachability. | `C/design.md:46–48,58–60`; `exitloop.go:1223,1682–1685`; `runtime.go:420–433` | Restore explicit trigger conditions; document which signal covers healthy-but-held states and test that coverage. |

VERDICT: REJECT — assertion withdrawal removes the round-8 unsafe authority, but draft 10’s replacement human recovery argument leaves a P1 gap for healthy cycles that never produce acceptable terminal evidence.