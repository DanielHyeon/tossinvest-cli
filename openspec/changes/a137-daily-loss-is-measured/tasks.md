# a137 Tasks

Story: `STORY-TOS-a137` · **High-risk**(Guardian · 위험 입력) — FLM/BTM 선행 · Pre-Edit 선언 · 변이 원장 · 적대적 독립 리뷰 의무.

> 순서: a136 이 공유 production 파일(`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는 risk-management 요구)을
> 먼저 커밋한다. 그 전에는 새 파일만. base = a137 첫 Go 커밋의 부모.

## 0. 계약 · 증거

- [x] 0.1 proposal · design · spec 델타 — `f030e27e`
- [x] 0.2 AST 증거(분기 주장 전) — `analysis/ast-evidence/` 12 함수
- [x] 0.3 CodeGraph hard evidence — `analysis/code-context/`(codegraph 1.6.0, R1~R9)
- [x] 0.4 사용자 결정 Q1~Q8 · Q12 · Q13 기록(review.md §2) · D0 증거 독립 대조(ACCURATE-WITH-CORRECTIONS) 정정 반영(review.md §1)
- [ ] 0.5 **사람 승인 — Q5 를 위한 risk-management FX 요구 MODIFIED**(design D5 · review.md §3)
- [ ] 0.6 proposal-freeze 리뷰(4보이스) → review.md §4, 발견 반영
- [ ] 0.7 FLM/BTM — 완료: 첫 다리 `admit`(편집 전 좌표). a136 커밋 뒤: `RiskGuardian.PrecheckQFinalEntry` · `RiskGuardian.IssueEntry` ·
      `QFinalEntryIssuance` 운반 지점(`collectStrategyFirstLegAuthority` 계열) · 감독자 하류 확인(`runMarket` 은 ast-evidence 로 읽기만)
- [ ] 0.8 a136 공유 파일 커밋 확인(`git log -1 -- <file>` · `git diff`) → `capture_change_base.py` 로 base 고정 → Pre-Edit 선언(review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — 측정기 순수 함수 10 시험(`internal/dailyloss/measure_test.go`, 태그 `a137red`, `analysis/red/red-dailyloss.txt` — **미커밋**) ·
      원장 읽기(관리 · 조정 종결 · 성과 행 조인, 계좌 한정 질의) · Guardian 토큰(위조 · 만료 · 미래 · 일 키 불일치 · 부재) · `max(호출자, 측정)` ·
      측정 실패 = daily_loss 단 INPUT_UNAVAILABLE · kill switch 사유 보존 · 매도/손절은 측정 미호출 · 첫 다리 강화 + 원인 보존 · 체인 없는 `Issuer.IssueEntry` 생산 호출자 0 고정
- [ ] 1.2 GREEN — `internal/dailyloss`(Collect · Total) · 원장 읽기 새 파일 · Guardian `MeasureDailyLoss` · 강화 입구 새 파일 · precheck/IssueEntry 바닥 적용 ·
      첫 다리 분류 · riskcalc `RealizedPnL.AsOf` 주석 정정 · 일 키는 a136 공유 함수 호출
- [ ] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change`

## 2. 검증

- [ ] 2.1 `go vet` · `rtk proxy go test -count=1`(dailyloss · risk · riskcalc · execgw · journal · app/engine) · `-tags tossos_testseams` · `-race`(make test-race 대상) · `openspec validate --strict`
- [ ] 2.2 변이 원장(max → 호출자만 · 이익 상쇄 · 조정 종결 포함 · 성과 부재 무시 · 일 키 UTC · 토큰 신선도 무시 · 매도 경로 측정 호출 · 강화 누락 · 원인 버림)
- [ ] 2.3 적대적 독립 리뷰(분리된 컨텍스트) → `review.md`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a137-daily-loss-is-measured` + 독립 검증
