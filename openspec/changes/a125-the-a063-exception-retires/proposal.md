# a125 · a063 도 다른 change 와 같은 기준으로 판정한다

- **Feature**: `FEAT-TOS-001` — StockOS SDD toolchain adaptation
- **Story**: `STORY-TOS-a125`
- **Spec**: `sdd-workflow` (MODIFIED 1 · RENAMED 1 · REMOVED 3)
- **위험 등급**: Normal(거래 경로 무관). 다만 **게이트 판정을 직접 바꾸는** 편집이므로 무거운 규율을 적용한다
  (편집 전 Python FLM · RED 선행 · 변이 · 적대 보이스 1 + gstack) — `docs/WORKFLOW.md` 「비례 원칙」.
- **결정**: 사용자 일괄 승인 2026-09-28(a063 수리안 = 특례 폐기 → 일반 경로(재고정+귀속) 전환, 옵션 ②).
  Manager(Fable) 승인 2026-09-28.
- 작성 2026-09-28, 구현 팀메이트 초안 · Manager 검증.

## Why — 오늘 잰 것

a120(2026-09-06)은 a063 한 change 를 위해 비교 기준 특례를 만들었다. 특례의 구성은 다음과 같다.

- 고정 P(`da80ce31`)→E(`e65e394b`) 쌍, 6.2 MB 원장, detached·clean HEAD 요구.
- 창의 끝을 감사된 `source_commit` 으로 두기.
- 도구 22 자리(`check_analysis.py`), 모듈 하나(`execution_baseline.py` 523 줄), 시험 약 20 개(`test_check_analysis.py`)와
  모듈 시험 403 줄, 정본 스펙 SHALL 넷, WORKFLOW 절 둘, README 12 곳.

그 특례는 한 change 도 통과시키지 못했다. 이유는 둘이다.

1. **일반 작업 트리에서는 설 수 없다.** `resolve_base` 가 특례를 부르면 오늘 HEAD 에서
   `invalid execution-baseline adoption: adoption requires detached HEAD` 로 멈춘다(2026-09-27 활성 전수에서 a063 만 BASE-ERR).
2. **게이트가 영원히 특례 전용 판정을 낸다.** a063 의 남은 사람 운영 항목(tasks 4.2~4.4)과 무관하게 그렇다.

그 사이 저장소는 2026-09-27 에 base 재고정 **사람 절차**를 정본에 넣었다(`docs/WORKFLOW.md` 「사람 승인 base 재고정」,
a123 불구현 종결 `0beab6d8`). 그 절차는 일반 규칙 안에서 a063 의 창을 연다. 특례를 끄고 잰 값은 다음과 같다
(2026-09-28, HEAD `0beab6d8`).

| 항목 | 값 |
|---|---|
| a063 base(`da80ce31`) → 워킹트리 required | 340(증거 없음 331 · 해시 불일치 2) |
| base 뒤 a063 자기 Go 커밋(`_self_repair_commits`) | 1 — `676bd4b4` |
| `676bd4b4` 가 고친 기존 함수 | 9, **전부** a063 번들 있음(current 6 · base 3) |

즉 a063 은 재고정 절차의 a071 변형(자기 수정 전부 번들로 덮임, 선례 `77e36cca`)에 해당하는 후보다.

## What Changes

1. **특례 경로 제거.**
   - `check_analysis.py` 에서 이관 판정 · 이관 문맥 키(`execution_baseline_adoption` · `adoption_source`) ·
     감사 라벨 · 이관 전용 거절 문장을 없앤다.
   - `resolve_base` 는 모든 change 에서 `base-commit.txt` 의 커밋을 돌려준다.
   - `SDD_BASE_REF` 는 그 커밋과 같을 때만 받는다(오늘 일반 change 규칙 그대로).
2. **`execution_baseline.py` 와 `test_execution_baseline.py` 를 지운다.** 특례를 못 박던 시험은 지우지 않고 **반전**한다.
   예: "유효한 a063 기록이 있으면 E" → "유효한 a063 기록이 있어도 `base-commit.txt`".
   "이관 경로는 착지를 거절" → "a063 도 착지 규칙을 그대로 받는다".
3. **남은 기록은 무력하다.** 어느 change 디렉터리의 `execution-baseline.json` 도 게이트가 읽지 않는다. a063 의 것은
   삭제 커밋으로 지우고, 원장(`analysis/execution-baseline-ledger.json`)과 이관 리뷰 문서는 a063 의 역사로 둔다.
4. **문서.** WORKFLOW 「a063 legacy execution-baseline exception」 절과 착지 절의 언급을 지운다. 「이관 worktree 의
   외부 SDD 인터프리터」 절은 이관 서술을 걷어내고 일반 doctor 옵션으로 남긴다(`SDD_PYTHON` 기능 불변).
   README 12 곳도 고친다.
5. **a063 전환(사람 절차, a063 디렉터리).** 재고정 영수증을 만든다: 귀속 1 커밋 · 함수 9 · 번들 대조표 · 번들 신선도,
   base 번들 3 의 처분 근거. 승인 인용(사용자 일괄 + Manager)을 둔다. `base-commit.txt` 단독 커밋 뒤 `check_analysis` 프로브를 돌린다.
   **게이트가 열려도 a063 아카이브는 아니다** — tasks 4.2~4.4(사람 운영 승인)는 남는다.

## Non-goals

- `SDD_PYTHON` doctor 기능 · 착지 규칙(a122 K2 · 규칙 1~8) · 빌림 규칙 · 재고정 절차 자체는 바꾸지 않는다.
- a063 의 운영 항목(4.2~4.4) · 엔진 · 주문 경로 · Go 코드는 건드리지 않는다.
- base 재기록을 막는 새 가드는 만들지 않는다(a123 종결의 결론 — 사람 절차).

## Impact

- `tools/logic-map/check_analysis.py`, `tools/logic-map/test_check_analysis.py`, `tools/logic-map/README.md`.
  `tools/logic-map/execution_baseline.py` · `test_execution_baseline.py` 는 삭제한다.
- `fixture_git_env.py` 의 설명 한 줄을 고친다(시험 모듈 둘 → 하나).
- `docs/WORKFLOW.md`, `openspec/specs/sdd-workflow/spec.md`(아카이브 때 적용).
- `openspec/changes/a063-align-attestation-renewal-profile/`: `execution-baseline.json` 삭제, `base-commit.txt` 재고정,
  재고정 영수증(`review.md`).
- 창 영향 전수: `execution-baseline.json` 을 가진 change 는 활성·아카이브 전체에서 a063 하나다(2026-09-28 `find`).
  다른 change 의 판정은 불변이어야 한다 — tasks 1.1 이 전수로 잰다.
