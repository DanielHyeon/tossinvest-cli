# Proposal-freeze review: align-full-sdd-pm-contract

Date: 2026-07-31
Base commit: `1dbef864038c81f0a2982a03e7d9549369d21669`
Verdict: **ACCEPT WITH REQUIRED CORRECTIONS**

## Review composition

The proposal, design, delta spec, PM validator, PM tests, StockOS Full SDD reference,
and TossOS workflow were reviewed sequentially through Manager/CEO, adversarial
engineering, and developer-experience lenses. Repository policy for this session
prohibits spawning separate agents, so no independent-agent result is represented as
having occurred. Strict OpenSpec validation passed after correcting requirement
sentences so the normative keyword is present in the parser's first line.

## Findings and decisions

| ID | Severity | Finding | Decision |
|---|---|---|---|
| R1 | P0 | Written 1:1 policy is false in enforcement: 32 pre-existing active changes bypass Story mapping through the registry allowlist | Accept. Backfill exactly one Story for every one, cover this change with STORY-TOS-002, then delete the bypass |
| R2 | P1 | This change makes the active total 33; proposal wording could be read as covering only 32 | Clarify that 32 means pre-existing changes and STORY-TOS-002 covers the current change |
| R3 | P1 | Manual Story `status` can disagree with active/archive evidence | Replace with `intent` and derive `designed`, `in_progress`, `implemented`, or `archived` deterministically |
| R4 | P1 | A Story path must remain verifiable after archive adds a date prefix | Validate the declared active path exactly or the actual date-prefixed archive directory for the same change ID |
| R5 | P1 | Removing the allowlist before backfill would break all SDD gates | Backfill hierarchy and Stories first; switch validator and registry atomically in the same logical unit |
| R6 | P1 | Copying StockOS text mechanically would corrupt TossOS safety and tooling facts | Preserve the explicit TossOS list in design and verify it in the WORKFLOW diff |
| R7 | P2 | UI/design review is not applicable because this change has no operator UI or runtime route | No design phase artifacts required |
| R8 | P2 | Generated trackers can conceal stale manual state | Render only derived status and fail `--check` when generated files drift |

## Architecture and failure modes

```text
portfolio source (intent + hierarchy + openspec mapping)
              |
              v
PM validator -----> active OpenSpec directories
      |             archived OpenSpec directories
      |             tasks.md checkbox evidence
      v
derived Story state
      |
      v
generated trackers (read-only views)
```

| Failure mode | Prevention/test |
|---|---|
| Active change has no Story | Exact set coverage test |
| Two Stories claim one change | Duplicate reverse-map test |
| Story declares wrong path | Active/archive path validation test |
| Registry reintroduces bootstrap bypass | Registry-key rejection test |
| Manual lifecycle claim returns | Manual `status` rejection test |
| Archive transition leaves stale tracker status | Derived archive-state test and generated drift check |
| TossOS project rules disappear during alignment | Preservation checklist in design, tasks, and final diff review |

## Scope decision

In scope: PM hierarchy, Story mappings, PM validator/tests/generated views,
`docs/WORKFLOW.md`, and this OpenSpec contract.

Not in scope: production trading code, operator runtime toggles, broker behavior,
Guardian values, journal writes, VPN exposure, containers, and retroactive Stories
for already archived historical changes.

There are no unresolved taste decisions or user challenges. The user's stated
premises already settle the two important choices: StockOS is the methodology
reference, and TossOS-specific project behavior must remain unchanged.

## RED/GREEN evidence

- RED: `python3 -m unittest test_generate_master_tracker.py` produced 5 assertion
  failures and 1 error against the old flat `change_id`/manual `status`/allowlist
  implementation. The failures covered current-repository coverage, duplicate mapping,
  invalid path, manual status, stale rendering, and derived lifecycle.
- GREEN: the same command runs 9 tests successfully after the PM migration and
  validator update.
- Coverage audit: 33 active OpenSpec changes, 33 active Story mappings, 0 missing,
  0 extra, 0 duplicate.
- Function Logic Map: `not-applicable`. The implementation surface is the Python PM
  validator and Markdown/JSON portfolio data; TossOS Function Logic Map tooling is a
  Go AST extractor. Branch behavior is pinned by the 9 PM unit tests instead.

## Post-implementation review

Date: 2026-07-31
Verdict: **IMPLEMENTATION APPROVED; LANDING BLOCKED**

The final review checked the PM source registry and every generated view, active and
archive path resolution, Story lifecycle transitions, the WORKFLOW preservation list,
strict OpenSpec validation, and whitespace/diff integrity.

The generic Python quality analyzer reported the pre-existing monolithic shape of
`validate` and many false-positive “magic numbers” from PM IDs in tests. No security,
resource, exception, concurrency, or correctness finding was produced. A validator
refactor is not required to enforce the contract and would widen this governance
change; the explicit helper functions and branch tests are retained as the smaller
implementation.

## SDD fingerprint investigation

Symptom: an early `make sdd-check` reported a stale CodeGraph fingerprint immediately
after `make sdd-sync`.

Root cause: the long-running sync command had yielded a live process session while
CodeGraphContext/GBrain work continued; the check was started before that session was
polled to completion, so it correctly read the previous fingerprint.

Resolution: no repository code change. The live sync sessions were polled to exit 0,
which recorded the current worktree fingerprint. A subsequent `make sdd-check`
confirmed the CodeGraph hard-evidence match. GBrain was owned by another live project
process and remained an advisory busy warning, exactly as `docs/WORKFLOW.md` specifies.

## Landing gate blocker

`make gate CHANGE=align-full-sdd-pm-contract` passed tasks-file, unchecked-task, and
review-file checks, then stopped at Function Logic Map discovery. The persisted base
commit predates unrelated uncommitted Go work already present in the shared worktree,
so the gate attributed about 80 modified existing Go functions in console, engine,
config, exit-policy, and journal packages to this documentation/PM change.

Those Go functions are not in this Story scope and many already belong to other active
OpenSpec changes. Stashing or committing the user's work, changing the base to conceal
it, or copying another change's Function Logic Maps into this change would violate the
single-writer and evidence-ownership rules. Task 5.3 therefore remains unchecked.
This change must be gated in a clean worktree containing only its own diff, or after
the owning changes have landed and this change's base has been legitimately rebased.
Until then it must remain active and must not be archived or reported as Full SDD
complete.

## 종결 시퀀스 (2026-09-27, a122·a067·a068 선례 · Manager 2단계 승인과 귀속 결정)

### Function Logic Map: not-applicable

**사유.** 이 change 의 범위에 Go 가 없다 — proposal Impact: "production trading code, runtime config,
주문·위험·원장 동작: 영향 없음", 편집 대상은 `docs/WORKFLOW.md` · PM portfolio · `tools/pm/`(Python) ·
`openspec/specs/sdd-workflow` 뿐이다. 번들 0 이 맞다.

### base 재고정 영수증 (4d413cf1)

- 잰 순간: HEAD f098105b, 모집단 = `git log --full-history -- openspec/changes/align-full-sdd-pm-contract` 커밋 2.
  - c0619279(2026-07-31, "ship automated portfolio operations console") — 5 change 를 한 커밋에 담은 squash, 886 파일,
    go 52 · `changed_existing_functions(c0619279^, c0619279)` = 53.
    **c0619279 의 align 몫 13파일에 .go 0 (docs/WORKFLOW.md · change dir 10 · tools/pm 2), proposal Impact 는
    '영향 없음' — 그 커밋의 기존 함수 수정 53건은 같은 squash 에 든 형제 4 change 몫**
    (add-common-exit-optimization 166 · enable-engine-autostart-menu 314 · enable-vpn-console-access 267 ·
    console-adoption-controls 9 파일; 넷 다 아카이브됨 — 2026-08-29 셋, 2026-07-31 하나).
  - 7b0dd0c9(2026-09-09, Full SDD 계약 복원 문서) go 0
- 옛 base 1dbef864 에서 5단계 rc 1 · required 546 · 창에 착지 커밋 604 — 전부 다른 change 의 기존 함수
  (`/tmp/claude-1000/a122-lot/audit/ca-align-full-sdd-pm-contract.log`).
- 재고정 1dbef864 → f098105b, `base-commit.txt` 만 커밋(4d413cf1). 재고정 뒤 check_analysis:
  required 0, 남은 사유는 이 마커뿐.
- 위 절이 적은 조건 — "after the owning changes have landed and this change's base has been legitimately
  rebased" — 이 성립했다: 그 Go 를 소유한 넷이 모두 아카이브됐고, 재고정은 사용자 결정(a122, base history 는
  git 이 정본)과 Manager 귀속 결정에 따른 것이다. 남의 작업을 숨기거나 남의 FLM 을 복사하지 않았다.
