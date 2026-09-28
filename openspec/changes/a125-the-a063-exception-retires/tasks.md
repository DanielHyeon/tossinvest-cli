## 0. 계약과 증거

- [x] 0.1 `base-commit.txt` 고정(`capture_change_base.py`) · Story `STORY-TOS-a125` 등록 · tracker `--check`
- [x] 0.2 `openspec validate a125-the-a063-exception-retires --strict`
- [x] 0.3 편집 전 Python FLM — 편집 대상 함수 AST 열거(a122 `enumerate.py`) + 반전 대상 시험 목록(AST 로 셈)
- [x] 0.4 proposal-freeze 리뷰 → `review.md` (적대 보이스 1 + codex — 게이트 판정 변경)

## 1. RED

- [x] 1.1 창 영향 전수: 모집단(활성+아카이브 전 change 의 `execution-baseline.json` 보유) + 분기 논증(`analysis/1.1-population.md`) + 편집 전 표본 A/B(고정 워크트리 `a125-census`). 전 change 실측 A/B 는 하지 않는다 — 중량 change 한 건이 6~9 분(2026-09-28 부분 실측 6건)이라 두 판에 수 시간, 비례 원칙
- [x] 1.2 반전 시험(design D3)을 먼저 쓰고 오늘 코드에서 RED 확인

## 2. GREEN

- [x] 2.1 `check_analysis.py` 이관 분기 · 문맥 키 · 라벨 · 거절 문장 삭제, `resolve_base` 단순화
- [x] 2.2 `execution_baseline.py` · `test_execution_baseline.py` 삭제, `fixture_git_env.py` 설명 갱신
- [x] 2.3 README · WORKFLOW 갱신(이관 절 삭제, `SDD_PYTHON` 절 일반화)

## 3. VERIFY

- [x] 3.1 편집 후 표본 A/B — 같은 워크트리 · 같은 HEAD 에서 a063 외 판정 불변(차이 0), a063 은 P 기준으로 바뀜
- [x] 3.2 변이(사본 · 무변이 대조군): 이관 분기 복원 · 착지 거절 복원 · `_recording_refusal` 거절 복원 · 기록 `execution_base` 를 base 로 읽기 · id 로 착지 우회 · 비정규 기록 거절 복원 · `SDD_BASE_REF` 대조 대상 변이 → 반전 시험 빨강(freeze F9)
- [x] 3.3 편집 후 Python FLM 재추출 · `make sdd-test` · `make sdd-check` · 적대 보이스 1 + gstack 리뷰 — `make sdd-check` 는 공유 트리에서 탐침 시한(부하)으로 미통과, 게이트 ⑥ 이 격리 워크트리에서 강제한다(review 「3.3 · 4.x 상태」)

## 4. a063 전환 (사람 절차 — a063 디렉터리)

- [x] 4.1 a063 `execution-baseline.json` 삭제 커밋. a063 `review.md` 에 소급 고지(`retrospective-exception` · "missing original analysis remains debt")를 옮겨 적고, a063 `tasks.md` 4.0 에 "a125 에서 폐기" 주석(freeze F5)
- [x] 4.2 조건 ① 영수증(재고정 **아님** — Manager 결정 (나)): 귀속 커밋 · 오늘 측정 함수 9 · base 번들의 해시를 창이 요구하는 P 판본으로(F4 실측 정정 — revision 은 base 가 맞다) →
      특례 없는 도구로 옛 base P 에서 판정 — **양성 단언**: 판정이 끝까지 갔고(조기 오류 없음) 실제 P · HEAD · 창이 출력과 같으며,
      귀속 함수마다 `(소스, 함수, 번들)` 대응이 서고, 그 번들들의 **전체** 오류(번들 디렉터리명으로 나오는 해시 · revision ·
      분기 · 호출 · 시험 인용 포함)가 0 이다. 남는 오류는 전부 형제 함수 누락으로만 분류된다(codex F1). renumber · S 사본 사실(F7) ·
      승인 인용(사용자 일괄 2026-09-28 + Manager)
- [x] 4.3 a063 `review.md` 에 재고정 연기 사유(재고정 값은 게이트 시점의 사실 — 사람 운영 4.2~4.4 뒤에만 의미)와, a063 `tasks.md` 에
      **게이트 직전 재고정 항목**: P → 당시 HEAD 귀속 재실측 · 번들 재검증(4.2 와 같은 양성 단언) · 문서/Go 분리 커밋에 대한 사람 귀속
      확인 · 영수증 HEAD 와 재고정 부모가 다르면 재측정(codex F2). a063 `issues.md` I1 에 승인된 전환과 남은 운영 차단을 날짜로 덧붙임(codex F5)

## 5. 종결

- [ ] 5.1 tracker 재생성 → `make sdd-sync` → Manager 게이트 슬롯 → `make gate CHANGE=a125-…`(격리 워크트리) → Manager 검증 → archive · Story 경로
