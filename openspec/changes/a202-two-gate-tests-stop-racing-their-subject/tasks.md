# a202 tasks

> 착수 2026-11-25(base `7bc35e5d1`, 브랜치 `test/a202-gate-tests-stop-racing`). 출처: a168 archive `issues.md` I3 · I4. 사용자 승인 2026-11-25 「승인」.

## 1. 계약과 증거

- [x] 1.1 Story `STORY-TOS-a202` 를 registry · `FEAT-TOS-001` 에 등록하고 proposal · design · tasks 를 쓴다. spec delta 없음(`skip_specs: true`, design §5). `openspec validate <id> --strict` 통과.
- [x] 1.2 `base-commit.txt` 고정, 기억 회고(`scripts/memory-recall.sh` — 결과 0건), CodeGraph 증거와 a166 census 고정 여부를 `analysis/code-context/` 에 기록한다.
- [x] 1.3 편집 전 두 시험 함수의 FLM/BTM 번들(`tools/logic-map`)을 만들고 커밋한다.

## 2. 구현 (RED → GREEN)

- [x] 2.1 RED: 자식 모드 직접 실행(deadlock fatal · exit 2)과 사본 지연 실험(옛 시험 FAIL)을 `analysis/red/` 에 원문으로 남긴다.
- [ ] 2.2 GREEN: `TestChildLockAndCrash` 자식 대기(D1)와 `TestTheReplacedReaderIsClosed` 대기 조건(D2)을 고친다. 편집 뒤 AST 재추출 · 번들 갱신, `check_analysis.py` 통과.

## 3. 검증

- [ ] 3.1 design §3 표의 실험(자식 생존 · 사본 지연 GREEN · 양성 대조 FAIL · `-count=200` 각각)과 두 패키지 전체(무태그 · `tossos_testseams`), `make lint` 원문을 `analysis/green/` · `analysis/verify/` 에 남긴다.
- [ ] 3.2 census(design §4) 원문을 `analysis/census/` 에 남긴다.
- [ ] 3.3 `review.md` — gstack 리뷰 생략 사유, 변이 원장 · 다중 적대 리뷰 `not-applicable` 사유, 대조 실험이 변이 원장을 대신함.

## 4. 종결 (Manager · 사용자 지시 대기)

- [ ] 4.1 Manager 독립 검증.
- [ ] 4.2 완료 게이트(사용자 명시 요청 시) → archive → Story 경로 · PM sync.
