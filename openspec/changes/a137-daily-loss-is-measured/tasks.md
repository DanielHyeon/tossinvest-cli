# a137 Tasks

Story: `STORY-TOS-a137` · **High-risk**(Guardian · 위험 입력) — FLM/BTM 선행 · Pre-Edit 선언 · 변이 원장 · 적대적 독립 리뷰 의무.

> 순서: a136 이 공유 production 파일(`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는 risk-management 요구)을
> 먼저 커밋한다. 그 전에는 새 파일만. base = a137 첫 Go 커밋의 부모.

## 0. 계약 · 증거

- [x] 0.1 proposal · design · spec 델타 — `f030e27e`
- [x] 0.2 AST 증거(분기 주장 전) — `analysis/ast-evidence/` 12 함수
- [x] 0.3 CodeGraph hard evidence — `analysis/code-context/`(codegraph 1.6.0, R1~R9)
- [x] 0.4 사용자 결정 Q1~Q8 · Q12 · Q13 기록(review.md §2) · D0 증거 독립 대조(ACCURATE-WITH-CORRECTIONS) 정정 반영(review.md §1)
- [x] 0.5 사람 승인 — Q5 를 위한 risk-management FX 요구 MODIFIED: 2026-10-14 사용자 선택 「1」 승인(review.md §2 · §3). 조항의 자리는 a136 델타로 옮김 권장(§5, Manager 조정 대기)
- [x] 0.6 proposal-freeze 리뷰(4보이스) → review.md §4: APPROVE-WITH-FIXES(P1×5 · P2×6 · P3×4), 전부 문서 · RED 에 반영
- [x] 0.6.1 P1 다섯 닫힘의 독립 재확인 — **BLOCKED**(P1-3 미닫힘 + 새 (b) · (a) · (c) · P2 셋) → 문서 · RED 수리 반영(review.md §4.1)
- [x] 0.6.2 §4.1 수리의 재확인 — **FREEZE-OK**(review.md §4.2). 조건: P1-3 RED 다섯을 GREEN 전에 실패 관찰. P3 셋은 GREEN 에서 적용
- [x] 0.7 FLM/BTM — 편집 함수 11 번들(편집 전 ast-base.json + 편집 뒤 ast.json, check_analysis 계산 목록과 일치) · 무편집 근거 dispatch · runMarket 은 `analysis/ast-evidence/flm-unedited/`. 이전 기록: 완료(편집 전 좌표): `admit` · `dispatch` · `runMarket`(무편집 근거, P1-3 — 먼저 작성). a136 커밋 뒤: `RiskGuardian.PrecheckQFinalEntry` ·
      `RiskGuardian.IssueEntry` · 전달 루프(`deliverEachStrategyHandoff` · `dispatchStrategyMarketHandoffs`) · 첫 다리 조립(토큰 · 짝 FX 운반 지점)
- [x] 0.8 a136 공유 파일 커밋 확인(c8783ad9 · de9617ab · 3685583d · a000cf29) → base = `a000cf29`(이 change 첫 Go 커밋의 부모) → Pre-Edit(review.md). 원문: a136 공유 파일 커밋 확인(`git log -1 -- <file>` · `git diff`) → `capture_change_base.py` 로 base 고정 → Pre-Edit 선언(review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — `analysis/red/` 8 영수증(컴파일 RED 5 · 행동 RED 2 — freeze 조건 ① · 펜스). 완료분: 측정기 순수 함수 12 시험(`internal/dailyloss/measure_test.go`, 태그 `a137red`, `analysis/red/red-dailyloss.txt` — **미커밋**:
      손실만 · 계좌 일 키 귀속 · 늦은 기록 · 조정 종결 제외 · 엔진 SELL 섞인 조정 종결 측정 불가 · 성과 부재 · 해석 불가 · 환산 · 0 일).
      **freeze 조건(§4.2) — GREEN 전에 실패 관찰 필수**: `TestA137ADailyLossModeBlockEndsTheCycleQuietly` · `TestA137OtherModeBlocksBehaveAsBefore`(기동 복원 실패 포함) ·
      `TestA137AMixedCycleStillReportsItsSkippedScopes` · `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` · `TestA137OtherFencesBehaveAsBefore`.
      남은 RED: 원장 읽기(관리 · 조정 종결 decimal 0 · 수명 안 발송 엔진 SELL · 성과 조인 · 계좌 한정) · Guardian 토큰(위조 · 만료 · 미래 · 일 키 불일치 · 부재 ·
      봉인 오류) · 바닥 규칙(무효 호출자 값 대체 금지 · 측정 실패의 호출자 값 낙하 금지 · 둘 다 유효할 때 max) · kill switch 사유 보존 · 매도/손절 측정 미호출 ·
      첫 다리 강화 + cause + `Stage == StageChain` · 분류 불변 표 · 계좌 진입 차단 결과로 주기 끝(refreshOnly · effective 두 조립 모두 잠금 0) ·
      trigger 일일 손실 모드 차단만 조용 · 다른 trigger · 기동 복원 실패 · 다른 관문은 오늘과 같음(양방향) · 강화 경합 펜스 · 혼합 주기(건너뛴 범위 보고 유지) ·
      보호 주문 CREATE/REPLACE 가 있던 조정 종결 측정 불가 · 형식 결함 `closed_at` 행이 사전 필터를 지나도 측정 불가 ·
      짝 FX 검증 · US 환산 올림 · 체인 없는 `Issuer.IssueEntry` 생산 호출자 0 고정
- [x] 1.2 GREEN — 대상 시험 통과(`go test` execgw · dailyloss · risk · riskcalc · clock rc 0, `-tags tossos_testseams` execgw rc 0, engine a137 rc 0). P3 셋 적용(§4.2: 조용한 경로 = trigger 일일 손실 AND Actor AUTO · 형식 결함 = 정규 RFC3339 UTC `Z` 아님 · 보호 필터 = 수명과 겹침) · `internal/dailyloss`(Collect · Total) · 원장 읽기 새 파일 · Guardian `MeasureDailyLoss` · 강화 입구 새 파일 · precheck/IssueEntry 바닥 적용 ·
      첫 다리 분류 · riskcalc `RealizedPnL.AsOf` 주석 정정 · 일 키는 a136 공유 함수 호출
- [x] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change` — 추가 리팩터 없음. 편집 뒤 AST 는 a1cfe0c2 · c83241ff · 4ddb081b · bd723b7f 에서 재추출, 새 base 63fd0c92 기준 `check_analysis` rc 0(bd723b7f, required 9 — evidence complete)

## 2. 검증

- [x] 2.1 `analysis/verify/receipts.md`(연결 워크트리 4fc5d4a4: test 108 ok · seams 109 ok · make test-race · A137 race 전부 rc 0) — `go vet` · `rtk proxy go test -count=1`(dailyloss · risk · riskcalc · execgw · journal · app/engine) · `-tags tossos_testseams` · `-race`(make test-race 대상) · `openspec validate --strict`
- [x] 2.2 변이 원장 28/28 CAUGHT(`analysis/mutation/ledger.md`, 1회차 M20 SURVIVED → 시험 추가로 닫음) — 변이 원장(max → 호출자만 · 측정 실패 시 호출자 값 낙하 · 무효 호출자 값 대체 · 이익 상쇄 · 시장 현지일 귀속 · 조정 종결 무조건 제외 · 엔진 SELL 검사 삭제 ·
      성과 부재 무시 · 토큰 신선도/일 키 무시 · 매도 경로 측정 호출 · 강화 누락 · `Stage` 검사 삭제 · 원인 버림 · 차단 결과를 오류로 · 다른 관문 거절도 조용히 · 다른 trigger 모드 차단도 조용히 · 펜스 원인 무시 · 건너뛴 범위 보고 삭제 ·
      보호 주문 검사 삭제 · 형식 결함 행 검사 삭제 · 환산 내림)
- [x] 2.3a 구현 적대 리뷰(a1cfe0c2) APPROVE-WITH-FIXES — P1×3 · P2×3 · P3 수리(review.md §6) — `c83241ff` · `3d48c22b`, 전 스위트 · race · 변이 37/37(receipts.md)
- [x] 2.3b 수리 재확인(470324ce) APPROVE-WITH-FIXES — P1-N1 · P2-N1 · P3-1 · P3-2 수리(review.md §7)
- [ ] 2.3 적대적 독립 리뷰(분리된 컨텍스트) → `review.md`

- [x] 2.4 원장 읽기 지연 측정 — `analysis/verify/latency.md`(bd723b7f 사본: 5,000 행 6.1 ms p95 · 20,000 행 25.9 ms p95 · 50,000 행 70.3 ms p95, 전수 스캔 별도). 원문: 원장 읽기 지연 측정 — 현실 크기 픽스처(관리 포지션 수천 · 일 키 창 안 수십)에서 측정 질의 1회의 시간과 단일 연결 점유 시간.
      `countMalformedClosedAt` 은 계좌의 CLOSED 행 전체를 positionSelect 로 훑음(형식 결함은 범위 사전 필터로 못 거르므로) — 그 전수 스캔 시간을 따로 잴 것

## 3. 사람 · Manager

- [x] 2.5 사람 승인 base 재고정 `a000cf29` → `63fd0c92`(Manager 승인 2026-10-15, review.md §6) — 귀속: 자기 Go 커밋 a1cfe0c2 · 4fc5d4a4 의 11 함수 전부 fresh 번들, 옛 창 required 24 = a137 11 + a136 13(형제 착지 몫). 새 base 기준 수리 커밋 함수 5 번들 추가
- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a137-daily-loss-is-measured` + 독립 검증
- [ ] 3.2 archive — a136 archive 뒤
