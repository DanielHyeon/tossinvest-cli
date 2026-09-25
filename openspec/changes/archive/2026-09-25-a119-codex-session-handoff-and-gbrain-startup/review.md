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

---

# Re-freeze review (task 2.3, second pass) — 2026-09-25 — **PASS (frozen under scope (a))**

- 기준: HEAD `dc289a7a`(a119 문서는 `2fbdcd78` 상태), base `54004f44`, 초안 `wip/a119-3.1` (5e151b97).
- 보이스: 독립 적대 리뷰 서브에이전트 1개 — **Claude Sonnet**, 읽기 전용, 구현 컨텍스트와 분리. 1차(거부)와 같은 형태이고
  모델만 다르다: Opus 주간 한도(리셋 09-26 19:00) 중이라 Manager 가 **의도적으로 조기 실행**했다 — 도구 change 의
  WORKFLOW 최소치는 validate + Manager 셀프리뷰이고, Manager 가 재작성 저자라 셀프리뷰 대신 독립 보이스를 세웠다.
  위 "Scope decision" 절의 "after the Opus reset" 문장은 이 결정으로 대체된다. `[codex-unavailable]` 아님 — Codex 는
  이 change 에서 쓰지 않는다(1차와 같음).
- 판정: **PASS**. blocking 0 · should-fix 0 · note 3.

| # | 심각도 | 위치 | 발견 | 처분 |
| --- | --- | --- | --- | --- |
| 1 | note | review.md "Scope decision" | 재리뷰 시점이 "Opus 리셋 뒤"로 적혀 있는데 09-25 에 실행됨 | **수용** — 위 문단에 조기 실행 사유 기록 |
| 2 | note | proposal Follow-ups 3 / issues I-2 | 후속 3(SDD agent-save 핸들러)에 결정 축이 없음 | **수용** — issues I-2 에 한 줄 추가(matcher 를 Codex 실제 이름에 맞출지 · coexist 의도를 포기할지) |
| 3 | note | codex-session-save 델타 :4-7 | "추가 이름" 요구가 오늘은 공집합 | **조치 없음** — 1차 리뷰 4번이 "이름이 추가될 때 문다"로 이미 수용, 은폐된 전칭-공집합 아님 |

리뷰어가 확인해 깨끗하다고 보고한 것(인용 포함): 문서 일관성(proposal ↔ 두 델타 ↔ tasks 3.x ↔ design 계획 ↔ host-evidence) ·
미확립 호스트 주장 0(긍정문 grep) · 정본 `sdd-workflow` :162-206 GBrain 소유권·contention·복구 요구 보존 · 정본
`codex-session-save` 4개 요구 유지 · 세 pin 이 `wip/a119-3.1` 실제 코드와 일치 · design "Task 3.2 evidence map" 인용 줄 실재 ·
baseline 대비 설정 파일 5개 diff 0 · 뮤테이션 하네스가 사본만 변이하고 대조군 실패 시 중단 · 안전 불변식 자명.

**Manager 재검증(스팟체크)**: `wip/a119-3.1` 시험 수 5 + 5 = 10 (`grep -c 'def test_'`) · evidence map 인용 줄
`test_gbrain_project.py:62·:154`, `test_codex_session_save.py:76·:256` 이 각각 duplicate-serve busy · stale heartbeat ·
Codex-store-only · concurrent-no-overwrite 시험의 `def` 줄 — 주장과 일치 · `openspec validate a119 --strict` valid(Manager 도 실행) ·
a119 디렉터리 `git status` 빈 출력(리뷰어 쓰기 0).

**다음**: 3.1 은 별도 Opus 팀메이트(리셋 뒤). 3.3 은 관측이 아니라 "미관측 기록"이다(scope (a)).

---

# Implementation verification (task 3.2) — 2026-09-26

- Base `54004f44`, HEAD after 3.1 `c202b804`. Teammate (Opus), separate from the Manager context.
- All runs below were read from raw output (`rtk proxy`), not from summarized output.

| Requirement (design evidence map) | Test run in isolation | Result |
| --- | --- | --- |
| Lock-owner preservation | `tools/sdd/test_gbrain_project.py` `:62` duplicate-serve busy · `:93` exit releases without deleting lock · `:118` live legacy owner busy, lock kept · `:154` stale heartbeat left for recovery | 4 tests, RC=0 (whole file: 8 tests, RC=0) |
| Isolation, redaction, atomic persistence | `tools/sdd-history/test_codex_session_save.py` `:76` Codex store only · `:134` bounded text + redaction · `:217` malformed stdin fails open · `:226` atomic publish, 5 backups · `:256` concurrent no-wait/no-overwrite | 5 tests, RC=0 (whole file: 7 tests, RC=0) |
| Unchanged tool result | `test_codex_host_event_coverage.CodexSaverToolResultTests` — stdout empty on success (per fixture name) and on the failure path | 2 tests, RC=0 |

Isolation of the runs themselves: every test above works in a `tempfile` repository — the `gbrain` tests with a copied wrapper
and a fake `gbrain` binary, the saver tests by running the real `.codex/hooks/save_session.py` against a temporary
git repository (wording corrected after the task 3.4 review, finding A10). Measured around the run: the real project lock owner (`.sdd/gbrain-home/.gbrain/tossos-process.lock`,
`pid`/`command`) was identical before and after and the owner process was still alive; the real
`.codex-context/session-summary.md` mtime was unchanged.

Byte identity against `54004f44` (sha256 of `git show 54004f44:<f>`, `git show HEAD:<f>` and the working tree):

| File | Result |
| --- | --- |
| `.codex/hooks.json` · `.codex/config.toml` · `.mcp.json` · `save-session.sh` · `.codex/hooks/save_session.py` · `tools/sdd/gbrain_project.py` | base = HEAD = working tree for all six |

`git diff --stat 54004f44 -- .mcp.json .claude save-session.sh tools/sdd/gbrain_project.py .codex/hooks.json .codex/config.toml`
→ 0 lines; `git status --short` on the same paths plus `.codex` → 0 lines. Commits tagged `[a119` since the base touch only
`openspec/changes/a119-…/**`, the two new test files and the two new fixtures — no trading, account or Claude-owned path.

Function Logic Map: not-applicable — no existing Go or Python function body changed; the change adds two test modules,
two JSON fixtures and one harness script under the change directory (`analysis/code-context/evidence-reconciliation.md`).

---

# Post-implementation review (task 3.4) — 2026-09-26

- Scope: `c202b804` (3.1), `2fd6dfc3` (3.2), `d76d9e18` (3.3); base `54004f44`.
- Voices, each a separate read-only sub-agent context (Claude Opus), apart from the implementing teammate:
  - **A — independent adversarial review.** It re-ran every committed claim and ran its own mutations E1–E12 on temp copies,
    one at a time.
  - **G — gstack `review` pre-landing lens, as a substitute.** `autoplan`/`review` are interactive pipelines and cannot ask
    questions in this autonomous session. Voice G read `~/.agents/skills/gstack/review/checklist.md` and applied Pass 1
    (critical) and Pass 2 (informational). This follows the a114 precedent (`archive/2026-09-25-a114-…/review.md` §1).
  - Codex was not used; this is not `[codex-unavailable]`. Codex model calls are a side effect this change does not take (as in 2.3).
- Verdicts: A **PASS-with-fixes** (blocking 0 · should-fix 3 · note 8); G **PASS** (critical 0 · informational 4).
  Both reproduced the committed numbers and confirmed that neither wrote to the repository.

| id | sev | finding | disposition |
| --- | --- | --- | --- |
| A1 | should-fix | The saver has three exits. The lock-contention exit (`save_session.py` `if lock is None: return 0`) had no stdout pin; mutation E9 survived. The paths had been read by hand; logic-map is Go-only. | **Accepted.** New test `test_lock_contention_exit_leaves_stdout_empty` holds the lock with `fcntl` and asserts empty stdout and no summary written; harness M6 added. Round 2 (R1) showed that "no summary" alone does not prove the exit was taken (E13 diverts to the exception exit), so the test also asserts empty stderr. |
| A2 / G1 | should-fix / info | The anchoring test was lexical: `^Bash\|apply_patch$` (E2) and `^(Bash\|apply_patch\|.*)$` (E1) passed. The non-goal "no name not established by a fixture" had no pin. | **Accepted.** The anchoring test now probes `x<name>` and `<name>x` under `re.search`. New `test_matcher_admits_no_name_the_fixture_has_not_established` checks §2.2 model-side names and the SDD handler names; any name later established by the fixture drops out of the probes automatically. Harness M7 (`.*`) and M8 (precedence) added. |
| A3 | should-fix (doc) | "A PostToolUse hook's stdout is read by the host as a decision" was stated as fact without a source. | **Accepted.** `proposal.md`, `design.md` and the test docstring now call it a precaution, unobserved for Codex `async` hooks. |
| A4 / G2 | note / info | The registration classifier misses shell and launcher forms (`bash -lc`, `/usr/bin/env gbrain`, `python3 -m`). | **Recorded, no change.** Speculative. The realistic raw form (`$GBRAIN_BIN serve`) and `/usr/bin/env python3 <wrapper>` are caught. The pin covers the direct form only. |
| A5 | note | The raw-`gbrain` pin covers the project layer only; the user-global layer is not pinned; `outside_repository_wrapper_registrations` is never asserted. | **Recorded.** User-global config is outside the repository and this change (§3.1 counted 0 there on 2026-09-25). Not pinned and not carried to a follow-up (round 2 R2 corrected the earlier "left for follow-up 2"). |
| A6 | note | Fixture sanitization is partial: free-text host fields; `~`, backslashes and 8-character session prefixes pass; the registration fixture has no sanitization test. | **Recorded, no change.** Both fixtures are committed and reviewed; their content is clean (leak scan by A). A stricter schema is not planned. |
| A7 | note | In the per-name subTest loop, the file-exists assertion was vacuous after the first name. | **Accepted.** `.codex-context` is removed before each name. |
| A8 / G3 | note / info | The harness never asserted `testsRun > 0`, and it leaked a `/dev/null` handle. | **Accepted.** `failures()` returns the run count, the harness stops when it is 0, and it uses `with open(os.devnull)`. |
| A9 | note | U5 and U6 map to follow-up 2, but the proposal text did not carry them. | **Accepted.** One sentence added to proposal follow-up 2. |
| A10 | note | review.md 3.2 said "or a copied saver"; the saver tests run the real saver against a temp repo. | **Accepted.** Wording corrected in place. |
| A11 | note | `host-evidence.md:14` kept an absolute system path, against the document's own rule. | **Accepted.** Replaced with a description. |
| G4 | info | §4 and the 3.2 block are English inside mostly Korean documents. | **No change.** a119 documents are kept in English by instruction; the Korean parts date from 2.2/2.3. |

Proposal and design edits are wording only. They reflect review decisions, which the WORKFLOW exempts from re-review; no
spec-delta Requirement changed.

## Verification after the fixes (2026-09-26, raw output via `rtk proxy`)

| Command | Result |
| --- | --- |
| `python3 -m unittest tools/sdd-history/test_codex_host_event_coverage.py tools/sdd/test_codex_gbrain_registration.py` | 12 tests OK (7 + 5), RC=0 |
| `TMPDIR=<scratchpad> python3 -W default analysis/harness/mutate_codex_config.py` | M0 control 0 failing (ran 12); M1–M8 failing 1·4·3·4·4·1·12·2 (subTests counted one by one); SURVIVED none; RC=0; no ResourceWarning; real `.codex/*` sha256 and `.codex-context/` listing unchanged |
| `make sdd-test` | RC=0 — 15 · 452 (skip 1) · 76 · 29 · 16 · 18, go ok |
| `openspec validate a119-… --strict --no-interactive` (installed 1.4.1) | valid |
| `openspec validate --all --strict --no-interactive` | 58 passed, 0 failed |
| `python3 tools/sdd/check_agent_config_sync.py` | RC=0, synchronized |
| `make sdd-sync` (before the fixes, at `d76d9e18`) | RC=2. The CodeGraph step completed; the CodeGraphContext step failed on a kuzu `Could not set lock on file` (another process holds the DB; advisory, the same failure as in a114); GBrain was busy (owned by a live session) and kept its previous freshness |
| `make sdd-check` (right after) | RC=0 — the CodeGraph hard-evidence index matches the worktree; the CGC and GBrain advisory indexes are warned as stale |

**Open for the Manager — gate step 5 is expected to fail.** `python3 tools/logic-map/check_analysis.py --change a119-…`
returned RC=1 at `d76d9e18`. The window runs from base `54004f44` to the working tree and held 27 commits from other
changes when judged at `d76d9e18` (the count grows as parallel commits land; 30 at `9f4a7aa1`). It requires 8 Go functions that a119 never touched: `cmd/tossctl/console.go:runConsole`, two
`cmd/tossctl/console_test.go` tests, one `a108_publication_is_total_test.go` test, and
`internal/strategyprojectionrpc/transport_unix.go` `reclaimStaleControlDirectory`, `verifyStaleSocketShape`,
`projectionSocketAccepts` and `Dial` (a113/a114 work). a119 has no `revision: current` bundle, so `--record-landing`
cannot narrow the window. This is the same step-5 policy block that the pending archive candidates hit (base behind
later work, zero bundles). It is not fixable inside a119 without rewriting its base or borrowing other changes' evidence.
The Manager decides.

## Round 2 — verification of fix `24433404` (voice A, same read-only rules)

Verdict: **PASS-with-fixes** — blocking 0 · should-fix 0 · note 4.
- E1, E2 and E9 are now caught.
- The real matcher raises no false positive: the control ran 7 + 5 + 7 with 0 failing.
- The committed harness reproduced M0 = 0 (ran 12) and M1–M8 = 1·4·3·4·4·1·12·2.
- Every disposition was confirmed except A1.
- The lock test is not timing-dependent: it takes a non-blocking `flock` on its own handle.

| id | finding | disposition |
| --- | --- | --- |
| R1 | "No summary written" does not prove the contention exit was taken. E13 turns that exit into `raise OSError`: stdout stays empty, no summary is written, and only a stderr warning appears. E13 survived. | **Accepted.** The lock test also asserts `stderr == ""`, and the A1 wording is corrected. Harness M9 (= E13) added. Teammate check on a copy: E13 now fails `test_lock_contention_exit_leaves_stdout_empty`; the control ran 7 with 0 failing. |
| R2 | A5 said the user-global gap was "left for follow-up 2", but follow-up 2 does not carry it. | **Accepted.** A5 reworded: not pinned, not carried. |
| R3 | A6 said "a follow-up" without naming one. | **Accepted.** Reworded to "not planned". |
| R4 | The step-5 commit count drifts (27 at `d76d9e18`, 30 at `9f4a7aa1`), and the sdd-sync/sdd-check rows predate the fix. | **Accepted.** The count is dated. Both commands were re-run after the fixes (below). |

Still surviving and recorded: E3, E4, E5, E11 (launcher/shell forms, A4) and E10 (an extra handler in the saver group; the
stdout pins cover the saver only).

## Verification after round 2 (2026-09-26)

| Command | Result |
| --- | --- |
| `python3 -m unittest tools/sdd-history/test_codex_host_event_coverage.py tools/sdd/test_codex_gbrain_registration.py` | 12 tests OK, RC=0 |
| harness (`TMPDIR=<scratchpad>`, `-W default`) | M0 0 failing (ran 12); M1–M9 failing 1·4·3·4·4·1·12·2·1; SURVIVED none; RC=0; real `.codex/*` sha256 and `.codex-context/` listing unchanged |
| `make sdd-sync` | RC=2. The CodeGraph step completed. CodeGraphContext hit the kuzu lock again (advisory). GBrain was busy with a live owner and kept its previous freshness |
| `make sdd-check` (right after) | RC=0 — the CodeGraph hard-evidence index matches the worktree. `sdd-test` inside it: 15 · 452 (skip 1) · 76 · 29 · 16 · 18, go ok |
| `openspec validate a119-… --strict` / `--all --strict` (1.4.1) | valid / 58 passed, 0 failed |

The fingerprint above covers the worktree *before* this record was written. Parallel sessions keep committing, so the
Manager re-runs `make sdd-sync` and `make sdd-check` right before `make gate`.
Task 3.4 stays **unchecked**: the final gate and Manager acceptance are still to come.

## Manager acceptance (2026-09-26, tossos-5c)

- Independent re-verification: the 12 new tests RC=0; config-surface diff vs `54004f44` is 0 lines; strict validate valid; **no `.go` file appears in any a119 commit** (`git show --stat c202b804 2fd6dfc3 d76d9e18 24433404 51249b5c` — measured, 0 hits outside openspec/).
- **Base re-pinned to `4798d3992c95f8dcd0fdf7faf812203910bbf9e3`** (the a119 2.3 commit, parent of the first implementation commit `c202b804`). Reason: the old base `54004f44` predated a113/a108 landings, so gate step 5 demanded 8 functions **other changes** edited; a119 edits no Go (measured above), and the new base still precedes every a119 implementation commit — this is the a114/a115 re-pin pattern, not a base-after-work waiver. With the new base `check_analysis` reports `required 0 function(s) … evidence complete or diff-proven exempt`, RC=0.
- Function Logic Map: not-applicable — no existing Go/Python function body changed by this change (the harness and tests are new files; measured above).
- Gate runs in an isolated worktree pinned at this change's completion commit, per house procedure.
