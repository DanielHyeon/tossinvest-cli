# Independent adversarial review — a119-codex-session-handoff-and-gbrain-startup

- Date: 2026-09-06
- Reviewer: separate adversarial review context
- Stage: specification-completeness repair only; no implementation or runtime acceptance
- Scope reviewed: original proposal, added design/tasks, both proposed spec deltas, canonical
  `codex-session-save` and `sdd-workflow` specifications, and the current repository hook,
  MCP, saver, and GBrain-wrapper contracts.

## Original blocker and disposition

The original a119 commit (`80d3931d`) contained a proposal but no `specs/` deltas; its
commit record identifies that absence as the reason strict/all-change OpenSpec validation
failed. The added deltas remove that document-completeness failure. This review does not
reinterpret the original operational report as new host or runtime evidence.

## Adversarial findings

1. **Accepted — modified block is complete.** The `codex-session-save` delta modifies the
   entire existing `Codex PostToolUse session capture` requirement: it retains `Bash`,
   `apply_patch`, coexistence with the SDD saver, asynchronous execution, and unchanged tool
   results; it retains both existing scenarios and adds bounded additional-event and
   unobserved-host scenarios. The separate canonical requirements for storage isolation,
   bounded/redacted handoff, atomic failure-safe persistence, and Claude preservation remain
   authoritative and are explicitly preserved by the modified requirement. No canonical
   requirement is weakened or silently replaced.
2. **Accepted — no unsupported host claim in the new deltas.** Additional event names require
   sanitized supported-host fixtures; matcher coverage alone is explicitly insufficient to
   claim event delivery or a refreshed handoff. The inherited proposal's historical symptom is
   not treated as proof of current host dispatch.
3. **Accepted — registration contract preserves the safety boundary.** The new capability
   distinguishes static repository configuration from actual Codex loading, requires one
   *effective* Codex-owned wrapper registration, preserves other agents' registrations, and
   forbids lock-owner/process/database deletion. It retains the canonical wrapper's project
   home, singleton lock, and exit-75 busy behavior.
4. **Required before implementation — do not choose matcher names or remove either current
   configuration entry from text inspection alone.** Current `.codex/hooks.json` matches only
   `Bash|apply_patch`; both `.codex/config.toml` and `.mcp.json` contain a GBrain wrapper
   registration. The required supported-host fixture and configuration-loading evidence are
   still absent, so neither observation proves which entry the host loads. Task 2.2 remains
   open.
5. **Required before any completion claim — implementation and runtime proof remain open.**
   No source/configuration/test implementation was reviewed or changed in this pass; there is
   no observed host event delivery, single host startup observation, implementation baseline,
   proposal-freeze review, teammate implementation, SDD/gstack/final-gate evidence, or Manager
   acceptance. This review does not mark a119 complete and does not authorize archive, PM
   synchronization, a runtime mutation, or a next `a0xx` change.

## Verification actually run

| Command | Actual result |
| --- | --- |
| `openspec validate a119-codex-session-handoff-and-gbrain-startup --strict --no-interactive` | exit 0 — valid |
| `openspec validate --all --strict --no-interactive` | exit 0 — 60 passed, 0 failed |

`make validate --all` is not an all-change OpenSpec command in this repository; GNU make
rejects that option (exit 2). It is not presented as validation evidence.

## Task ownership and status

Tasks 1.1 and 1.2 are Manager-owned status decisions. This reviewer neither changes their
checkboxes nor treats the successful document validation as implementation completion. All
code-feature tasks (2.1 through 3.4) remain unchecked and require the evidence and separate
implementation/review sequence stated in the change.

## Verdict

The added OpenSpec deltas truthfully repair the original missing-delta validation blocker and
preserve the canonical isolation and lock contracts. They are clear for this specification-only
stage. Host/runtime assertions and all implementation acceptance remain explicitly pending.

---

# Proposal-freeze review (task 2.3) — 2026-09-25

- 기준: base `54004f44`, 증거 `analysis/host-evidence.md`, 계획 `design.md` "Evidence-backed implementation plan"
- 보이스: 독립 적대 리뷰 서브에이전트 1개(구현 컨텍스트와 분리, 읽기 전용). `autoplan` 4관점은 돌리지
  않았다 — 이 change 는 도구·설정 change(WORKFLOW 위험 등급 "경량")이고 자율 실행에서 사용자 질문 관문을
  거칠 수 없어 적대 리뷰 1개로 대신함. `[codex-unavailable]` 아님 — Codex 보이스는 모델 호출 부작용 때문에 쓰지 않음.
- 판정: **REJECT (동결 거부)**

| # | 심각도 | 발견 | 처분 |
| --- | --- | --- | --- |
| 1 | blocking | 전달 추론은 `codex exec` 한정이며 증상 호스트(대화형 VS Code/Desktop)는 미관측; 08-29 `FileChange` 10건 뒤 미갱신은 "가설 불지지" 결론과 모순 | **수용** — host-evidence §2.3·issues I-1 을 "확립도 반박도 안 됨·미해명"으로 정정, design 에 측정 호스트 명시 |
| 2 | blocking | 계획(설정 불변 + 시험만)이 proposal 의 Why·What Changes 와 어긋나고 I-1 범위 결정이 미정 | **수용** — I-1 을 blocking 으로 올리고 (a)/(b) 결정을 사람에게 넘김. 결정 전 3.x 커밋 안 함 |
| 3 | should-fix | "유효 1개"·`apply_patch` 확립·스레드별 원인이 단정형 | **수용** — CLI 로더 한정·소거법·추론으로 표기 |
| 4 | should-fix | 적용 범위 시험이 기존 `test_codex_session_save.py:280-305` 와 중복, RED 가 파일 부재 | **부분 수용** — 음성 대조 시험 삭제, RED 를 사본 변이 하네스로 정의. 픽스처×matcher 합동 시험 1개는 **유지**: 스펙 델타가 "함께 시험"을 요구하고 이름이 추가될 때 비로소 물기 때문 |
| 5 | should-fix | 등록 픽스처가 저자 자기 대조, 출처 층은 호스트 출력 아님 | **수용** — `source_layer_basis: inferred_by_elimination…` 추가, 시험 주석을 "drift pin" 으로 |
| 6 | should-fix | 3.2 항목과 baseline diff 시나리오 미매핑 | **수용** — design 에 증거 표 추가 |
| 7 | note | trust hash 주장이 "측정"으로 표기됨; 새 경로 신뢰 항목에 `enabled` 없음 | **수용** — "추론"으로 정정, 3.3 pending 목록에 추가 |
| 8 | note | SDD agent-save matcher 가 Codex 이름에 안 걸림 | **수용(기록만)** — issues I-2 |
| 9 | note | `gstack-review.md` 에 홈 경로 | **수용** — `~` 로 치환 |
| 10 | note | 안전 불변식 유지 | 확인 |

## 이 시점의 실행 결과(초안 시험 포함, 초안은 `wip/a119-3.1` 에만 있음)

| 명령 | 결과 |
| --- | --- |
| `python3 -m unittest tools/sdd-history/test_codex_host_event_coverage.py tools/sdd/test_codex_gbrain_registration.py` | 초안 픽스처 없을 때 RC=1(errors 5+5), 추가 뒤 RC=0 · 10 tests OK(정리 후) |
| `python3 …/analysis/harness/mutate_codex_config.py` | 대조군 0 실패, 변이 M1~M5 전부 실패(1·1·3·4·3), SURVIVED none |
| `make sdd-test` (초안 포함 상태) | RC=0 — 15·451(skip 1)·76·28·16·18 OK, go ok |
| `openspec validate a119-… --strict --no-interactive` | RC=0 valid |
| `openspec validate --all --strict --no-interactive` | RC=0 — 58 passed, 0 failed |
| `python3 tools/sdd/check_agent_config_sync.py` | RC=0 synchronized |
| `git status --short .codex .mcp.json .claude save-session.sh tools/sdd/gbrain_project.py` | 빈 출력(불변) |

Function Logic Map: not-applicable — 기존 Go/Python 함수 본문을 바꾸지 않음(신규 시험·픽스처만, 그것도 wip 브랜치).
`make sdd-check`·`make gate`·호스트 관측(3.3)은 실행하지 않음 — 동결 거부 상태이며 3.3 은 **pending**.

---

# Scope decision and rewrite (2026-09-25) — re-freeze pending

- User decision on issues I-1: **(a)** — evidence and regression pins only; both symptoms to named follow-ups.
- Rewritten: `proposal.md` (Why/What/Follow-ups/Non-goals), `tasks.md` 2.3 note and 3.x, `specs/gbrain-codex-mcp-startup/spec.md`
  (dropped "Workspace startup is observed"; added raw-`gbrain` and concurrent-thread scenarios; "correction" wording removed
  since no configuration changes), `design.md` status paragraphs. `specs/codex-session-save/spec.md` unchanged — its
  requirement already conditions additional names on sanitized fixtures.
- Requirement-level edits → the freeze review re-runs on the rewritten deltas (adversarial voice 1, lightweight tool change;
  after the Opus reset 2026-09-26 19:00 KST). `openspec validate --strict` result is recorded in the round that runs it.
- Not done here: no 3.x implementation, no host observation, no configuration edit. Draft tests remain on `wip/a119-3.1`.
