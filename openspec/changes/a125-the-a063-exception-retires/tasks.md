## 0. 계약과 증거

- [x] 0.1 `base-commit.txt` 고정(`capture_change_base.py`) · Story `STORY-TOS-a125` 등록 · tracker `--check`
- [x] 0.2 `openspec validate a125-the-a063-exception-retires --strict`
- [ ] 0.3 편집 전 Python FLM — 편집 대상 함수 AST 열거(a122 `enumerate.py`) + 반전 대상 시험 목록(AST 로 셈)
- [ ] 0.4 proposal-freeze 리뷰 → `review.md` (적대 보이스 1 + codex — 게이트 판정 변경)

## 1. RED

- [ ] 1.1 창 영향 전수: 활성+아카이브 전 change 의 편집 전 판정(rc · required · 대상 라벨) 기록 — 사본 하네스, 대조군 선행
- [ ] 1.2 반전 시험(design D3)을 먼저 쓰고 오늘 코드에서 RED 확인

## 2. GREEN

- [ ] 2.1 `check_analysis.py` 이관 분기 · 문맥 키 · 라벨 · 거절 문장 삭제, `resolve_base` 단순화
- [ ] 2.2 `execution_baseline.py` · `test_execution_baseline.py` 삭제, `fixture_git_env.py` 설명 갱신
- [ ] 2.3 README · WORKFLOW 갱신(이관 절 삭제, `SDD_PYTHON` 절 일반화)

## 3. VERIFY

- [ ] 3.1 창 영향 전수 편집 후 — a063 외 판정 불변(차이 0), a063 은 P 기준으로 바뀜
- [ ] 3.2 변이(사본 · 무변이 대조군): 이관 분기 복원 · 착지 거절 복원 · `SDD_BASE_REF` 대조 대상 변이 → 반전 시험 빨강
- [ ] 3.3 편집 후 Python FLM 재추출 · `make sdd-test` · `make sdd-check` · 적대 보이스 1 + gstack 리뷰

## 4. a063 전환 (사람 절차 — a063 디렉터리)

- [ ] 4.1 a063 `execution-baseline.json` 삭제 커밋(기록은 a063 `review.md`)
- [ ] 4.2 재고정 영수증: 귀속 커밋 · 함수 9 · 번들 대조표 · 신선도 · base 번들 3 처분 근거 · 승인 인용(사용자 일괄 2026-09-28 + Manager)
- [ ] 4.3 `base-commit.txt` 단독 커밋 → `check_analysis` 프로브 결과 기록. a063 tasks 4.2~4.4(사람 운영)는 남는다고 명기

## 5. 종결

- [ ] 5.1 tracker 재생성 → `make sdd-sync` → Manager 게이트 슬롯 → `make gate CHANGE=a125-…`(격리 워크트리) → Manager 검증 → archive · Story 경로
