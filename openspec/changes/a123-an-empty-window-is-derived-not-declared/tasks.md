## 0. 계약과 증거

- [ ] 0.1 `base-commit.txt` 고정 — **a122 아카이브 뒤**, proposal-freeze 직전 (`capture_change_base.py`)
- [ ] 0.2 `openspec validate a123-an-empty-window-is-derived-not-declared --strict` 통과
- [ ] 0.3 Python FLM **편집 전**: `_landing_refusal` · `compute_landing` · `_self_repair_commits` · `check` ·
      `changed_existing_functions` (a122 `enumerate.py`) → `analysis/python-function-logic/`
- [ ] 0.4 proposal-freeze 리뷰 → `review.md` (도구 change 경량 + 적대 보이스 1 — 게이트가 받는 것을 늘리는 편집)

## 1. RED — census 와 위조 픽스처

- [ ] 1.1 사본에 R1·R2 를 넣고 활성+아카이브 전수 「같음/획득/상실」 표 — 상실 0, 획득 = proposal 표
- [ ] 1.2 위조 픽스처 (a) base 만 옮긴 커밋 거절 · (b) 위조 번들 → K2 경로 · (c) 디렉터리 안 만진 Go 커밋 → 요구 누락(못 박기만) · (d) 병합 전용 번들 → 창 전체
- [ ] 1.3 뮤테이션: 조건 1·2·3 각각 제거 → 픽스처 빨강 (사본 · 무변이 대조군 선행)

## 2. GREEN

- [ ] 2.1 R1 — 세 조건 판정과 빈 창 유도
- [ ] 2.2 R2 — 번들 0 change 의 귀속 요구 집합
- [ ] 2.3 R3 — 판정 줄 출력(재기준화 sha · 귀속 커밋 · 요구 수), 골든 픽스처

## 3. VERIFY

- [ ] 3.1 수용 측정: a077 · a079 · a074 5단계 PASS / a067 · a068 "자기 함수 N" (N 실측) / align 사유 출력
- [ ] 3.2 `README.md` 갱신 · `test_check_analysis.py` 통과 · `make sdd-test` · `make sdd-check`
- [ ] 3.3 편집 후 Python FLM 재추출 · gstack 리뷰 · Manager 검증

## 4. 종결

- [ ] 4.1 `make gate CHANGE=a123-…` · archive · Story 경로 · PM `--check`
- [ ] 4.2 align 사람 재기준화 요청(별도, 선례 `840b3377`)
