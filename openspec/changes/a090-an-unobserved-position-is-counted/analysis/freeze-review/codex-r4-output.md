# codex 4라운드(좁은 확인) 출력 원문 (a090 0.6f)

- session `01a0e9da-a230-7903-b6ec-54b94e113f8c`, codex-cli 0.154.0, gpt-6-astra, read-only, 트리 `git archive 0b5ea307`
- 2026-09-29 06:10:14~06:12:02 KST, rc 0, tokens 110,307, 401 없음. 프롬프트 `codex-r4-prompt.md`(485933e8, sha256 6adede68…)
- 아래는 `-o` 로 받은 마지막 메시지 그대로다(편집 없음).

---

Read-only review completed; no writes, tests, network calls, or engine execution. Status below concerns **proposal resolution**, not implementation.

References: `A` = `openspec/changes/a090-an-unobserved-position-is-counted`; `B` = `openspec/changes/a092-an-alert-does-not-hold-the-stop`; `E` = `internal/app/engine/exitloop.go`.

| Finding | Status | Confirmation |
|---|---|---|
| R3-1 | **PARTIAL** | RED keys and GREEN routing are repaired (`A/tasks.md:45–48,96–102`), but contradictory direct-enqueue/account-field instructions remain; see N1. |
| R3-2 | **RESOLVED** | Dedicated a090 adapter replaces a092 announcer reuse, removes `FieldAccount`, selects transition-ID key and window zero (`A/design.md:124–130`; `A/tasks.md:100–102`). |
| R3-3 | **PARTIAL** | D13 explicitly accepts shared-component account/error logging as the Manager’s separate issue; it does **not** eliminate the original exposure. Failure canaries establish only the narrower producer-input claim (`A/design.md:216–223`; `A/tasks.md:91–92`); see N2. |
| R3-4 | **RESOLVED** | Implementation is blocked until a092’s entrance lands; no operative fallback lifecycle remains (`A/design.md:94–98`; `A/tasks.md:24`). Stale instructions remain, but cannot override that explicit gate. |
| R3-5 | **RESOLVED** | Retries are limited to eligible processing cycles; pending announcements survive streak termination, with removal/B3 and skipped-cycle coverage specified (`A/design.md:236–240`; `A/tasks.md:59–64`). |
| R3-6 | **RESOLVED** | Ancestor conditions, occurrence ordinal and exit kind distinguish every frozen exit; multiplicity and unmatched-exit rejection remain required (`A/tasks.md:79–85`; `E:508–601`). |
| R3-7 | **RESOLVED** | New anchor-aware fixture must establish zero elapsed at creation and exact progression for pre/post-rollback anchors (`A/design.md:76–80`; `A/tasks.md:55–58`). |
| R3-8 | **RESOLVED** | Tracer B7 reachability and unchanged seven `o.alert` callers are corrected (`A/tasks.md:34`; `A/analysis/function-logic/internal-app-engine--exitobserver.observeonce/function-logic-map.md:61–62`; `A/analysis/code-context/codegraph-baseline.md:10`). |

The specified **alert and announcement inputs** are account-free under the selected production construction: position IDs and automatically generated transition IDs are hashes, not embedded account strings (`internal/journal/position_projection.go:441–446`; `internal/journal/operating_mode.go:720–726`). Automatic announcement actor/cause are fixed by `EscalateOperatingMode` (`internal/journal/operating_mode.go:504–509`), so removing the account field and replacing the key covers the account-bearing parts of the extracted event (`internal/obs/mode.go:62–71`). New dedicated logs have an explicit field allowlist (`A/design.md:164`).

That does **not** establish that every resulting shared-path log is account-free. D13 expressly excludes those logs, and task 2.17 cannot prove a broader claim.

| ID | Severity | Finding | Evidence | Fix |
|---|---|---|---|---|
| N1 | **P2** | Fifth-draft routing cleanup is incomplete. Hard dependency is clear, but active text still directs direct enqueue, mentions the deleted outside-entrance alternative, and lists account ref among alert fields. | `A/design.md:122`; `A/proposal.md:57`; `A/tasks.md:53,61,134`, versus `:24,97,101`. | Replace stale operational instructions with entrance-based wording; remove account ref from the safety summary. |
| N2 | **P2** | Failure canary is narrower than its surrounding privacy claim. On mode-commit failure, no announcement reaches the entrance, so checking entrance arguments can pass vacuously. It also does not require inspecting a090-owned failure-log output or injecting an account-bearing error. | `A/tasks.md:88–92`; commit errors precede announcement at `internal/journal/operating_mode.go:391–395,468–480`; logger serializes raw errors at `internal/obs/log.go:162–165`. | Assert each injected failure was reached; inspect a090-owned emitted logs and complete event contents, using account/error canaries. Preserve the expressly accepted shared-component exclusion. |
| N3 | **P2** | D14 records a defensible compatibility interpretation, but the upstream SHALL still literally covers all exit-goroutine records. The acknowledged erratum remains outstanding. | `A/design.md:205–212`; `B/specs/engine-safety/spec.md:44`; explicit a090/window-zero compatibility at `B/design.md:759–770`. | Apply the recorded clarification in a092’s implementation lot and retain a window-zero regression assertion. |

The census now has unique coordinates for all ten frozen exits. In particular, B12 and B21 have different ancestors—`result.Corruption != nil` versus `identityErr != nil`—despite sharing `qerr != nil` (`E:542–549,589–596`). D14 introduces no new runtime safety defect under the recorded Manager interpretation.

VERDICT: PASS — no new P0/P1 remains within the accepted Manager scope; privacy-test coverage and stale contract wording remain P2 corrections.