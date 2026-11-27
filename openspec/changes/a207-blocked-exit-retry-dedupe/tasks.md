# a207 tasks

> 사용자 목표: 막힌 청산 재시도가 반복 기록 · 반복 알림을 만들지 않게 함. 재시도 빈도 · 즉시성은 바꾸지 않음(§0.3).
> 1단계(이 커밋)는 조사 + 제안만이며 생산 코드 변경 0. 2단계는 proposal 「사람이 결정할 것」이 정해지고 READY 가 된 뒤 시작함. High-risk(청산 · 원장).

## 1. 조사 · 계약(1단계)

- [x] 1.1 기억 회고: `scripts/memory-recall.sh`(결과 0) · Hindsight `tossos-dev-episodes`(직접 선례 0) · ai-memory(세션 기록뿐 — 결정 페이지 0)
- [x] 1.2 반복 경로 AST 열거 21 개(`analysis/ast-evidence/claims.md` C1~C12)
- [x] 1.3 쓰기 · 읽기 · 알림 지점 조정: CodeGraph 1.6.0 callers · `affected --filter '*_test.go'` · `rg`(`analysis/code-context/evidence-reconciliation.md`)
- [x] 1.4 운영 저널 읽기 전용 사본 집계(`analysis/a207_measure.py` · `.out.txt`) · 압축 판정 질의 비용(`analysis/a207_latency.py`), 사본 삭제
- [x] 1.5 proposal · design · exit-policy spec 델타, Story `STORY-TOS-a207` → FEAT-TOS-009, `openspec validate --strict` · PM `--check`
- [x] 1.6 경량 리뷰 1회(`review.md`)

## 2. 구현(2단계 — READY 뒤)

- [ ] 2.1 사람 결정 1~5 반영, `base-commit.txt` 고정, Pre-Edit 선언(review.md)
- [ ] 2.2 FLM 번들: `Journal.recordExitJudgementTx` · `appendExitEventTx` · `hydrateExitEventEvidence` · `ExitObserver.release` · `ExitObserver.submit` · `ExitObserver.endExitOrderCycle`
- [ ] 2.3 RED: design §7.1 T1~T8
- [ ] 2.4 GREEN: D1~D4(journal 압축 · 리더 · 질의 실패 시 전체 행) → D5(release 메모리 셈) → D6(해소 알림)
- [ ] 2.5 변이 원장 M1~M12(design §7.2)
- [ ] 2.6 선택 시험 · `make lint` · `check_analysis.py --change a207-blocked-exit-retry-dedupe`
- [ ] 2.7 다중 적대 리뷰(작성자와 다른 계열 보이스 둘, 최대 2 회차)

## 3. 사람 승인 뒤(이 change 밖에서 사람이 시작)

- [ ] 3.1 main 병합 · `make image` · 엔진 교체(두 시장 닫힌 창)
- [ ] 3.2 운영 저널 읽기 전용 확인: 다음 막힘에서 압축 행 · 해소 알림 실측
- [ ] 3.3 archive
