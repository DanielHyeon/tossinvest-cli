# a130 review

## Pre-Edit Gate (2026-10-13, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a130-an-adopted-trade-keeps-its-outcome / 1.2
- 대상 심볼(패키지.함수): journal.freezeTradeOutcomeTx · journal.computeTradeOutcome · journal.adoptedRoundTripLegs ·
  journal.Journal.BackfillTradeOutcome · journal.ApplyExitFill · journal.Journal.runApplyHooks · engine.bindApplyHooks ·
  engine.NewContext (+ 타입 journal.ApplyHooks · journal.ApplyTx 필드 하나씩, 새 파일 journal/trade_outcome_refusal.go ·
  journal/trade_outcome_backfill.go · engine/trade_outcome_refusal_log.go · tools/a130-outcome-backfill — 새 leaf,
  console/templates_portfolio.go 상수 문자열)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/)
  — freeze 생산 호출자 1(ApplyExitFill) · BackfillTradeOutcome 생산 호출자 0 · bindApplyHooks 생산 1(NewContext) · 시험 14
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
  (R1: ApplyExitFill 은 함수 값 — CodeGraph 가 동적 호출을 못 봄, HEAD 로 해소)
- 기존 동작 파악 근거: base 233d5d9c 직접 읽기 · AST 번들 8 · 운영 저널 사본(읽기 전용) 재생 · 기존 시험
  trade_outcomes_test.go(15) · adoption_outcome_test.go(6) · adjustment_close_test.go · exit_wiring_internal_test.go ·
  interlock_internal_test.go · portfolio_test.go:752
- Function Logic Map / Branch Test Map: analysis/function-logic/ 8 번들(편집 전, 분기 1/12/5/6/15/8/1/15)
- upstream 상속 테스트 영향: no — journal 성과 · 엔진 기동은 TossOS 고유. journal · engine · console 전체 go test 로 회귀 확인
- 실패 테스트 선행 작성: yes — analysis/red/ (production 무편집 base 에서)
- 설정·DB·journal 변경과 rollback: 스키마 무변경(35). 원장 쓰기 동작은 「거절이던 편입 종결이 행이 됨」 한 방향.
  롤백 = 이미지 핀 되돌림(이미 쓴 행은 불변 트리거로 남음 — 정상 행). backfill 은 사람 실행.
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문 · 손절 · 판정 무접촉, 성과 실패는 여전히 체결을 되돌리지 않음, 거절 줄은
  트랜잭션 밖 goroutine(쓰기 잠금 보유 중 I/O 없음), 계좌 참조 · 포지션 id 미기록, 토글 flip 없음, 운영 저널 쓰기 0(에이전트)
```

## Proposal-freeze review (2026-10-13)

- 보이스: 적대적 Eng 1(분리된 리뷰 컨텍스트, code-reviewer, 읽기 전용 — 종료 시 `git status` 로 무편집 확인, 스크래치 비어 있음).
  대상: proposal · design 1판 · spec 1판 · tasks · FLM/BTM 8 · code-context. AST 번들 8개 `source_sha256` = base 파일, 분기 · return 수 = FLM 표 확인.
- 판정: **REJECT** (P0 2 · P1 4 · P2 4 · P3 3). 구현 착수 전이라 production 편집 0(초안은 트리 밖 보존 — issues.md).
- 「R 산식(분자 · 분모 같은 Q) 2배 부풀림 없음」 · 「양의 조정만의 시각 오분류는 fail-closed」 · 「로그 싱크 설계(recover · goroutine · nil 싱크 · 필드 넷 · 로거에 계좌 With 없음)」 은 성립 확인.

| # | 등급 | 발견 | 실측 · 처리 |
|---|---|---|---|
| F-1 | **P0** | `new_avg_price` 는 대사 수량 수렴 경로에서 브로커 평균가가 아님 — `internal/reconcile/converge.go:221` 이 항상 `NewAvgPrice: ""` 를 보내고 `position_adjustments.go` 가 `firstNonEmpty(req, target.AvgPrice)` 로 **프로젝션의 기존 평균가(편입 전 fold-in 때의 cost_basis)** 를 저장. 실측 11건 `prev = new` 인 이유. D1 대로면 늘어난 수량을 cost_basis 로 매입 처리 — A7 round-3 의 SHALL NOT(cost_basis 는 산식에 안 들어감) 위반, 불변 행이라 정정 불가 | **수용 · BLOCKED** — HEAD 로 확인(converge.go:221 · position_adjustments.go `firstNonEmpty`). 사용자 결정 1항의 전제(「그 대사 조정의 브로커 평균가」)가 사실과 다름 → **사람 결정 필요**(issues.md I-1) |
| F-2 | **P0** | backfill 모집단 「CLOSED 전부」 는 금지된 행을 만듦 — 엔진 진입 경로 `roundTripLegs` 는 매도 수량 검사가 없어(:211 매수만 · :425 priced 만) ADJUSTMENT_CLOSED(사람이 판) 포지션에 「매도 0 · priced」 → 전액 손실 조작 행. `id >= low` 창(:365)은 상한이 없어 닫힌 트랜잭션 밖에선 뒤 인스턴스의 체결을 쓸어 담음 | **수용** — 재설계: 모집단을 편입 포지션 · 엔진 매도 종결(ADJUSTMENT_CLOSED 없음)로 좁히고 spec 에 명시, 트랜잭션 밖 계산은 `committed_at <= closed_at` 상한 또는 뒤 인스턴스 존재 시 거절 |
| F-3 | P1 | 편입 뒤 음(−)의 조정 + 심볼 단위 flatten 쓸어 담기(F)로 Q_sell = q_a + ΣΔ⁺ − H + F 가 F = H 이면 Q_buy 와 같아짐(fail-open) | **수용** — 편입 뒤(`>=`) 음의 조정이 하나라도 있으면 거절(`post_adoption_disposal`), RED 시험(이전 인스턴스 flatten + 부분 수동 처분) |
| F-4 | P1 | 재편입(`resetExitStateForReadoptTx`, apply_hook.go:699 ← position_policy.go:145)이 `initial_risk` · `entry_price` 를 다시 씀 — 「exit 상태의 최초 위험 불변」 전제와 spec SHALL 이 재편입 포지션에서 거짓 | **수용** — HEAD 확인. 재편입 포지션은 거절(`readopted`) 또는 사람 결정(issues.md I-2) |
| F-5 | P1 | 규칙 5: 설계가 근거로 쓴 함수 중 AST 번들 없는 것 — `ApplyPositionAdjustment`(F-1 의 출처) · `reconcile.ConvergeQuantities` · `resetExitStateForReadoptTx` · `engineSellOrders` · `roundTripLegs` | **수용** — 재동결 전 번들 추가 |
| F-6 | P1 | 불변식 8: 2.3 이 운영 사본의 거래별 재계산(종목 · 수량 · 손익 · R)을 저장소에 씀 · design 사실절에 실제 종목 · 수량 | **수용** — 검증 산출물은 개수 · 일치 여부 · 사유 코드만, design 사실절 익명화(이번 판에서 466100 예 삭제) |
| F-7 | P2 | 실행 TOCTOU — 계획 재계산 → 토큰 비교 → BackfillTradeOutcome 이 세 번째로 계산해 autocommit INSERT, 쓴 값 ≠ 토큰이 검증한 값 가능 | **수용** — 항목마다 짧은 `BEGIN IMMEDIATE`(재계산 = 계획 값 대조 → INSERT), 첫 차이에서 멈춤. 스키마 동일 가드 · 계획 단일 읽기 트랜잭션을 design/spec 에 |
| F-8 | P2 | D3 의 줄이 커밋 전에 예약됨 — 롤백 · 재시도에서 거짓 · 중복 줄 | **수용** — 거절을 ApplyTx 에 모아 커밋 성공 뒤 호출자가 싱크로(goroutine 유지) |
| F-9 | P2 | 회귀 미명명: 편입 3 → +3 → −3 → 엔진 3 매도는 a130 이전엔 동결, D1 에선 거절(Q_buy 6) | **수용** — F-3 거절 규칙과 함께 D6 에 명명 · 시험 · 2.3 에서 31행 중 해당 수 세기 |
| F-10 | P2 | spec 충돌: 「성과 원시 지표」 의 「CLOSED 트랜잭션에서 동결 · 이후 비동기 작업이 다시 읽거나 갱신하지 않는다」 와 backfill 요구가 조정되지 않음 | **수용** — MODIFIED 절 또는 명시 예외, 사유 코드 집합 · 모집단 · 미마이그레이션을 델타에 |
| F-11 | P3 | 「싱크 nil → 읽기 동일」 은 편입 종결에선 거짓(조정 읽기는 싱크와 무관하게 돎) | **수용** — 문장을 싱크 경로로 한정 |
| F-12 | P3 | 시계 출처: observed_at = 드라이버 시계, created_at = 저널 시계, 재편입은 RFC3339Nano — 시계 어긋남 + 쓸어 담기 잔여 | **수용** — D6 잔여로 명명, spec 에 「position_id 명시 귀속, 시각은 한 인스턴스 두 사건의 순서만」 |
| F-13 | P3 | 시험 · 변이 보강(같은 초 경계 · 음의 조정 · 재편입 · 백필 모집단 제외 · 쓴 값 = 미리본 값 · 커밋 뒤 줄 · `-wal/-shm` 미생성), 콘솔 규칙 13 브라우저 확인 | **수용** |
