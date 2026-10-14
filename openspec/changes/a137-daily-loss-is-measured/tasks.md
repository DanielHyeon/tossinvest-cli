# a137 Tasks

Story: `STORY-TOS-a137` · **High-risk**(Guardian · 위험 입력) — FLM/BTM 선행 · Pre-Edit 선언 · 변이 원장 · 적대적 독립 리뷰 의무.

## 0. 계약 · 증거

- [x] 0.1 proposal · design(D0 현재 동작 증거 · D1~D6 권장안 · fail-closed 거부 목록 · 사람 결정 Q1~Q8) · spec 델타
- [x] 0.2 AST 증거(분기 주장 전) — `analysis/ast-evidence/` 10 함수(tools/logic-map, HEAD `7b073f55`)
- [x] 0.3 CodeGraph hard evidence — `analysis/code-context/`(codegraph 1.6.0, R1~R9, 불일치 R4 = 도구 한계 · R9 = 미고정)
- [ ] 0.4 **사람 결정 Q1~Q8**(design.md 「사람 결정」) — 결정을 design.md 에 옮기고 권장과 다른 답이면 spec 델타 · 거부 목록을 고친다
- [ ] 0.5 a136 착지 확인 → `capture_change_base.py --change a137-daily-loss-is-measured` 로 base 고정(공유 표면: `riskguardian.go` · Guardian policy · risk-management spec)
- [ ] 0.6 proposal-freeze 리뷰(autoplan 4관점 + 적대적 Eng) → `review.md` · CodeGraphContext 질의 갱신
- [ ] 0.7 FLM/BTM — `RiskGuardian.IssueEntry` · `RiskGuardian.PrecheckQFinalEntry` · (Q6) 첫 다리 거절 분류 · 감독자 `latchMarket` 하류(R9) · 주입 지점 · Pre-Edit 선언

## 1. 구현 (RED → GREEN → REFACTOR) — 0.4~0.7 뒤

- [ ] 1.1 RED — 서명값 0 · 원장 손실 ≥ 한도 → 거절 + 강화 / 호출자 값이 크면 그 값 / 성과 행 없는 오늘의 청산 → daily_loss 단 INPUT_UNAVAILABLE /
      측정 실패 중 위험 축소는 측정기 미호출 · 지연 없음 / kill switch 사유 보존 / 시장 현지일 경계 / 외화 환산 실패
- [ ] 1.2 GREEN — 원장 측정기(새 파일) · Guardian 진입 입구 둘의 `max(호출자, 측정)` · 생산 조립 주입 · (Q6) 전략 경로 분류 · 강화
- [ ] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change`

## 2. 검증

- [ ] 2.1 `go vet` · `rtk proxy go test -count=1`(risk · riskcalc · execgw · journal · app/engine) · `-tags tossos_testseams` · `-race`(make test-race 대상) · `openspec validate --strict`
- [ ] 2.2 변이 원장(max → 호출자만 · 측정 실패 무시 · 매도 경로 측정 호출 · 일 경계 UTC · 강화 누락)
- [ ] 2.3 적대적 독립 리뷰(분리된 컨텍스트) → `review.md`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a137-daily-loss-is-measured` + 독립 검증
