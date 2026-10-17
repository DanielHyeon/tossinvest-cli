# codex 10라운드(좁은 확인) 출력 원문 (a094 0.5u)

- session `01a0e9f8-f673-7d91-8db4-11efea4fe980`, codex-cli 0.154.0, gpt-6-astra, read-only, 트리 `git archive bc53211f`(17537)
- 2026-09-29 06:43:21~06:45:24 KST, rc 0, tokens 98,872, 401 없음. 프롬프트 `codex-r10-prompt.md`(88e05526, sha256 799e921b…)
- 아래는 `-o` 로 받은 마지막 메시지 그대로다(편집 없음).

---

Read-only review completed; no tests, network calls, or engine execution. `C/` = `openspec/changes/a094-a-stop-clears-what-blocks-it/`; `A/` = `openspec/changes/a090-an-unobserved-position-is-counted/`. References use this tree.

| Round-9 finding | Status | Reason |
|---|---|---|
| R9-1 | **RESOLVED** for the approved escalation scope | Healthy-but-nonterminal evidence now has an explicit critical trigger and OPEN/refused-CLOSED acceptance cases; unavailable recovery is explicitly deferred, not represented as working (`C/design.md:46–63`; `C/tasks.md:240–243`). |
| R9-2 | **RESOLVED** | D−9.1 distinguishes complete collection from per-snapshot commits, rejects success⇒terminality, and acknowledges the counterexample to the universal assertion claim (`C/design.md:21–27`). |
| R9-3 | **RESOLVED** | Existing alerts now have explicit conditions; D−2.7 is correctly described as planned, and detector degradation requires cycle failures (`C/design.md:31–37`; `internal/app/engine/runtime.go:420–433`). |

**D−9.3 adversarial assessment**

- **Journal-only evidence: implementable.** Pending intent, confirmed order ownership, CANCEL target/state/ID/settlement time, and accepted terminal snapshots are durable facts. No broker read or detector-health query is required; elapsed time additionally requires the clock (`internal/journal/apply_hook.go:572–582`; `internal/journal/schema.go:223–234`; `internal/journal/fills.go:1550–1567`; `C/design.md:46–48`). Use the already-required scoped terminal-evidence predicate; absence from `TrackedFillOrders` alone is insufficient.
- **Episode key: finite and restart-stable.** One key per position/cancel attempt, without timestamp or cycle counter. Attempt ID is a primary key. Window zero reuses existing rows and does not rearm delivered/acknowledged rows (`C/design.md:49–50`; `internal/journal/schema.go:224`; `internal/journal/outbox.go:289–294,346,382–384`).
- **Healthy detection, including refused CLOSED: covered.** Refusal preserves trusted snapshot state and returns nil error; the detector records fail-closed status but subsequently resets outage health. Consequently the missing-terminal predicate remains true despite healthy detection (`internal/journal/fills.go:384–409`; `internal/filldetect/detect.go:324–347`; `internal/brokerstate/derive.go:567–573`). Evaluation at **judge entry**, before suppression and early returns, is essential; D−9.3 specifies that placement (`C/design.md:46`; `internal/app/engine/exitloop.go:857–885`).
- **Unjudged positions: conditional cover, not universal.** a090’s target-set-minus-judged definition covers missing/unusable quotes and working-set filtering (`A/design.md:25–26`). But persistent working-set errors and repeated restarts remain explicit holes (`A/design.md:249–255`). D−9.3 overstates the union; finding below.
- **Spam/absorption:** repeated observations and ordinary restart cannot create new episode rows. The added suffix avoids collision with `noteDelay`’s `type|positionID` key (`internal/app/engine/exitloop.go:1686–1688`). Separate delay/degradation notifications may coexist intentionally. Suppression after delivery or acknowledgment is the declared one-shot policy. This establishes durable deduplication, **not exactly-once external delivery**: delivery settlement is a separate operation (`internal/obs/notifier.go:812–814`).
- **Wall-clock residual: not finitely bounded.** Repeated/backward clock adjustments can postpone eligibility indefinitely. D−9.3 names that limitation honestly, but 30 seconds is an eligibility threshold, not a guaranteed delivery deadline (`C/design.md:48,56`). Neither notification nor elapsed time authorizes release.

D−9.4 correctly identifies an unavailable recovery mechanism. Detector refusal maps to `ReasonBrokerStateUnknown` and blocks through the gate (`internal/filldetect/ledger.go:97–103`; `internal/filldetect/detect.go:669–682`). No matching hand-resolution command was found; existing `reconcile-resolve` handles quantity-mismatch RECONCILE blocks (`cmd/tossctl/engine_reconcile.go:79–89`). The named follow-up is a candidate, not a recovery guarantee.

**D−9.1 reasoning**

Accurate under the stated precedence. Application commits per snapshot (`internal/filldetect/detect.go:316–321`; `internal/journal/fills.go:541–553`), and successful cycles need not establish terminality. Withdrawal remains justified because the withdrawn mechanism cannot establish safe remaining quantity—not because truthful assertions can never change an outcome. Read “only authority” as the current accepted design’s authority, not a universal impossibility claim about future recovery designs.

| ID | Severity | Finding | Evidence | Fix |
|---|---|---|---|---|
| R10-1 | **P2** | New unconditional “union closes silence” wording exceeds a090’s expressly conditional coverage. This does not reopen a090’s accepted residuals. | `C/design.md:55`; `C/tasks.md:243`; `A/design.md:249–255` | Qualify the union by a090’s processing/restart conditions; explicitly inherit persistent-B2 and restart residuals. Keep the missing-quote acceptance case. |

VERDICT: PASS — no P0/P1 remains in this narrow escalation-and-wording review; record the P2 coverage overstatement, with recovery and notification timing explicitly unbounded.