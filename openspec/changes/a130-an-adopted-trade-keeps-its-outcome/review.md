# a130 review

## Pre-Edit Gate (4판, 2026-10-13, 구현 팀메이트 Opus — 1판 · 3판 선언을 대체)

```text
Pre-Edit Gate:
- change id / task id: a130-an-adopted-trade-keeps-its-outcome / 1.2
- 대상 심볼(패키지.함수): journal.freezeTradeOutcomeTx · journal.computeTradeOutcome · journal.adoptedRoundTripLegs ·
  journal.Journal.BackfillTradeOutcome · journal.ApplyExitFill · journal.Journal.runApplyHooks(핸들 리터럴 한 필드) ·
  journal.Journal.RecordFill(ctx 수집 목록 · 커밋 뒤 보고) · engine.bindApplyHooks(선택 인자) · engine.NewContext(바인딩 호출 한 줄)
  + 타입 journal.ApplyHooks · journal.ApplyTx 필드 하나씩, 새 파일 journal/trade_outcome_refusal.go · journal/trade_outcome_backfill.go ·
  engine/trade_outcome_refusal_log.go · official/a130_intake_candle.go(공식 1분봉 읽기 전용 GET) · tools/a130-outcome-backfill(새 leaf),
  console/templates_portfolio.go 상수 문자열,
  journal/adoption_static_test.go 패키지 변수(읽기 허용 목록 두 줄)
- 무편집 인용 근거: ApplyPositionAdjustment · Converger.ConvergeQuantities · resetExitStateForReadoptTx · engineSellOrders · roundTripLegs ·
  backfillConfirmedStrategyFillTx(번들 있음)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, raw/)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
- 기존 동작 파악 근거: base 233d5d9c 직접 읽기 · AST 번들 15 · 운영 저널 사본(읽기 전용) census · 관측 간격 실측 · 기존 시험
  trade_outcomes_test.go · adoption_outcome_test.go · adjustment_close_test.go · position_policy_test.go · exit_wiring_internal_test.go ·
  interlock_internal_test.go · position_campaign_hardening_test.go · adoption_static_test.go · readonly_test.go · portfolio_test.go:752
- Function Logic Map / Branch Test Map: analysis/function-logic/ 15 번들(편집 전)
- upstream 상속 테스트 영향: no — journal 성과 · 엔진 기동 · 콘솔 안내문은 TossOS 고유. journal · engine · console 전체 go test 로 회귀 확인
- 실패 테스트 선행 작성: yes — analysis/red/red-1.txt(컴파일) · red-2.txt(base 행동 RED: F-9 · F-3 조작 행 · 재편입) · red-4.txt(4판)
- 설정·DB·journal 변경과 rollback: 스키마 무변경(35). 원장 쓰기 동작: 체결 트랜잭션은 편입 뒤 증가(needs_intake_price) · 감소 · 재편입 ·
  인스턴스 밖 매도를 거절(보수), 증가 포지션의 행은 사람이 돌리는 backfill 만 씀. 공식 API: candles 1m 읽기 전용 GET(도구만, 1.2초 간격, 초당 5회 묶음 아래). 롤백 = 이미지 핀 되돌림(쓴 행은 불변 트리거로 남음). backfill 은 사람 실행(스키마 동일 가드).
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문 · 손절 · 판정 무접촉, 성과 실패는 체결을 되돌리지 않음(거절은 쌓기만 · 커밋 뒤 보고 · recover),
  체결 트랜잭션 안 네트워크 0, 로그 쓰기는 goroutine(체결 루프 비차단), 계좌 참조 · 포지션 id 미기록, 토글 flip 없음, 운영 저널 쓰기 0(에이전트), backfill 쓰기 잠금은 항목 하나 동안만
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

## 2판 처분 (2026-10-13, 사용자 결정 I-1 = (b) · I-2 = readopted 반영)

| # | 2판 처분 |
|---|---|
| F-1 | 해소 — 늘어난 수량을 가격 매기지 않음(`post_adoption_increase` 거절). 가격 대입 (a) 기각, converge 수리 (c) = a132 |
| F-2 | 해소 — backfill 도구 없음(D4: 구조상 0, 사본 census 로 확인). `roundTripLegs` 의 트랜잭션 밖 위험은 D5 잔여 |
| F-3 | 수용 — `post_adoption_disposal`(`>=`), 시험 `TestA130APostAdoptionDisposalBlocksASweptFlatten` |
| F-4 | 수용 — `readopted`(ExitEventReadopted 명시 기록), 시험 `TestA130AReadoptedPositionIsRefused` |
| F-5 | 수용 — 인용 근거 번들 5 + 편집 대상 추가 2(RecordFill · backfillConfirmedStrategyFillTx) |
| F-6 | 수용 — verify 는 수 · 사유만, design 사실절 익명화 |
| F-7 | 해소 — backfill 없음 |
| F-8 | 수용 — 핸들에 쌓고 RecordFill 커밋 성공 뒤 보고(D3), 롤백 무보고 시험 |
| F-9 | 수용 — 이름 붙인 회귀 `TestA130TheGrowThenShrinkRoundTripIsNowRefused`, D1 에 명명 |
| F-10 | 수용 — spec 「성과 동결 거절의 기록」 에 「성과 원시 지표 동결 규칙 불변」 명시(backfill 요구 삭제로 충돌 원인 제거) |
| F-11 | 수용 — 문장을 싱크 경로로 한정(proposal · D3) |
| F-12 | 수용 — 시계 출처 D1, 잔여 D5, spec 에 명시 귀속 문장 |
| F-13 | 수용 — D7 시험 · 변이 목록, 콘솔 브라우저 확인(D6) |

## 2차 proposal-freeze review (2026-10-13) — APPROVE-WITH-FIXES

- 보이스: 적대적 Eng 1(분리된 컨텍스트, code-reviewer, 읽기 전용 — 종료 시 `git status` 무편집 확인, 스크래치 비어 있음). 대상: 2판(리뷰 중 저자가 D3 · D4
  를 두 번 고침 — 리뷰는 고친 판 기준). AST 번들 15 `source_sha256` = base 파일 확인.
- 성립 확인: 2판의 「엄격히 더 거절」 · 재편입 판정 완결(호출자 하나 · 같은 트랜잭션 이벤트) · 수량 장부(q_a + d + b + F) · 시계 출처 · 커밋 자리(RecordFill
  B37/B38 하나, 전략 경로 BUY 만) · 엔진 싱크 비차단 · 손절 경로 무접촉.
- **그 직후 사용자가 2판 방침((b) 거절만)을 철회하고 재계산(묶음별 편입가)을 결정** → 3판. 2차 발견의 처분:

| # | 등급 | 발견 | 3판 처분 |
|---|---|---|---|
| R2-1 | P1 | 「성과 원시 지표」 와 델타가 모순(「동결 규칙 그대로」 는 거짓) · 재시도 시나리오 · Costs 주석 | **수용** — spec MODIFIED 「성과 원시 지표」(동결 조건 · 거절 · 복구 경로 명시), 「그대로」 문장 삭제. Costs 주석은 GREEN 에서 고침 |
| R2-2 | P2 | 거절 기록 SHALL 이 코드보다 넓음(:966 · :973 · 비 locallyOwned 는 시도 안 함) | **수용** — SHALL 을 「exit 상태가 열려 있던 포지션이 체결로 CLOSED 되어 동결을 시도한 경우」 로 좁힘, D2 표에 런타임 도달 여부 열 |
| R2-3 | P2 | Pre-Edit 이 1판 그대로 | **수용** — 3판 Pre-Edit 재선언(위) |
| R2-4 | P2 | census 가 트리 밖 코드 · 모집단 문구 불일치 | **수용** — census 시험을 커밋하고 GREEN 커밋 뒤 다시 재서 sha 기록(tasks 2.3), D4 모집단 문구 = census 의 backfill 모집단 |
| R2-5 | P2 | 기존 시험은 fold-in 시각 = 편입 시각 우연에 기대 green | **수용** — `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot`(30초 앞 · 같은 초), 새 시험은 시계를 명시적으로 진행, 변이 `>`→`>=` |
| R2-6 | P2 | sell_unpriced · 조정 없는 과귀속 · 자리별 사유 대조 시험 없음 | **수용** — `TestA130AnUnpricedEngineSellIsReportedAsSuch` · `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAMismatch` · `TestA130EachReachableRefusalSiteNamesItsOwnReason`, 변이 sell_unpriced→mismatch |
| R2-7 | P3 | D7 · FLM · tasks 의 낡은 문구, 가변 인자라 배선 삭제 변이가 컴파일됨 | **수용** — 문구 정리, NewContext AST 구조 핀이 싱크 인자를 잼 |
| R2-8 | P3 | adoption_static 허용 목록 넓히지 말 것(adoptedBasis 가 observed_at 반환) | **거절(명명)** — adoptedBasis(무편집 함수)를 고치면 FLM 대상이 늘어남. 허용 목록 두 줄은 그 시험이 요구하는 검토 경로이고 두 파일 모두 UPDATE 없음 — proposal Impact 에 명시 |
| R2-9 | P3 | 목록 없이 싱크만 있는 호출자는 거절을 조용히 버림 | **명명** — D3 · D5 잔여 |
| R2-10 | P3 | 일시 · 구조 실패가 섞인 사유 | **명명** — D5 |
| R2-11 | P3 | AdoptPosition 수량 미검사 · 감소 규칙의 시각 의존 | **명명** — D5. 감소 규칙을 시각 무관으로 바꾸는 대안은 편입 전 부분 처분을 거절하므로 채택 안 함 |

## 3차 proposal-freeze review (2026-10-13, 3판 621e2eb4) — REJECT

- 보이스: 적대적 Eng 1(분리된 컨텍스트, code-reviewer, 읽기 전용 — 종료 시 `git status` 무편집 확인, 스크래치 삭제). AST 인용 파일 8개 base = HEAD sha 동일 확인.
- 성립 확인: R 산식(분모 = 위험 × Q_buy = Q_sell, 2배 없음) · CLOSED 인스턴스에 조정 불가(새 인스턴스) · 종결 뒤 판정 기록 불가(ErrExitStateCompleted).
- 판정: **REJECT** (P0 1 · P1 2 · P2 6 · P3 5). 구현 착수 전(production 편집 0, 3판 초안은 트리 밖 · wip 브랜치).

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| R3-1 | **P0** | 결정한 가격 출처(조정 시각 뒤 엔진 관측)가 원장에 없음 — `exit_events` 는 exit 선이 바뀐 관측만 담음. 평평한 선은 `refreshObservation`(exitloop.go:1109) → `RefreshExitObservation`(exit_observation_refresh.go:37 「exit event 를 붙이지 않음」, 882a0b49 a111)으로 exit_states 만 덮어씀. 수량 증가는 이벤트를 강제하지 않음(ProjectedQuantity · SameExitOperationalLine · ChangedFromState 가 남은 수량을 안 봄). 그래서 p_k 는 「조정 뒤 첫 관측」 이 아니라 「조정 뒤 첫 **선 변경** 관측」(신고가 틱 또는 손절 무장 관측) — 손절 직전이면 손절가로 매겨져 손실이 0 쪽으로 당겨짐 | **수용 · BLOCKED** — HEAD 로 확인(exit_observation_refresh.go:37-38 · exitloop.go:1105-1114). 사용자 결정 2차의 전제가 원장 사실과 다름 → **사람 결정 필요**(issues.md I-3) |
| R3-2 | P1 | 383초는 관측 주기가 아니라 「선 변경 이벤트」 간격의 p99(모집단 · 방법 · 하네스 미커밋) | 수용 — I-3 결정에 따름(코드 상수 유도 또는 조정→첫 이벤트 분포 직접 측정, 스크립트 커밋) |
| R3-3 | P1 | 조정이 아닌 처분(엔진 제안 · flatten 이 아닌 로컬 SELL 체결)은 감소 규칙에 안 걸림 + 심볼 단위 flatten 쓸어 담기(F)로 상쇄되어 조작 행 가능 | 수용 — flatten 갈래를 인스턴스 창(opened_at, closed_at] 으로 묶고, 그 창의 귀속 밖 로컬 SELL 체결이 있으면 거절, RED 시험 |
| R3-4 | P2 | `journal.Open` 은 항상 이관 — 스키마 가드 순서 명시 · 32 파일 바이트 불변 시험 | 수용 — 초안 도구는 이미 OpenReadOnly → 버전 비교 → Open 순서. spec · D4 에 순서 명시, 시험 추가 |
| R3-5 | P2 | backfill 항목 잠금 시간 미측정 · 「분석 경로의 격리」 와 경합 | 수용 — 사본에서 항목 최대 잠금 시간 측정 · 실행 창(KST 05:00~09:00) 요구 또는 잠금 안 재대조를 지문(최대 id)으로 |
| R3-6 | P2 | MODIFIED 「성과 원시 지표」 자기모순(첫 문장 무조건 · CLOSED 트랜잭션 · 초기 수량 정의) | 수용 — 문장 정리 |
| R3-7 | P2 | 시각 정밀도 · 문자열 정렬 · READOPTED 의 observed_price · a042 이전 created_at 대체 | 수용 — RFC3339Nano 파싱 양쪽, 판정 행만(action 필터), NULL observed_at 은 거절, 소수 초 경계 시험 |
| R3-8 | P2 | FLM 규칙 5: `recordExitJudgementTx` · `appendExitEventTx` · `requiresFullJudgement` · `RefreshExitObservation` · `OpenAdoptedExitState` 번들 없음, R2-8 처분 문구 오류(adoptedBasis 는 observed_at 을 안 돌려줌) | 수용 |
| R3-9 | P2 | 시험 · 변이 보강(평평한 선 · 제안 이벤트 선택 · 소수 초 · NULL · READOPTED/OPENED 제외 · 로컬 SELL+flatten · 스키마 32 바이트 불변 / 문자열 비교 · created_at 대체 · 필터 · flatten 경계 · 지문) | 수용 |
| R3-10 | P3 | 거절 줄은 goroutine 이라 유실 가능 — spec 은 SHALL | 수용 — best-effort 명시 |
| R3-11 | P3 | base 233d5d9c 가 a131(f269c615) 이전 | 명명 — 게이트의 쌓인 change 검사에서 Manager 판단(재고정은 사람 절차) |
| R3-12 | P3 | census 시험이 사본을 Open 으로 열면 32→35 이관 | 확인 — census 는 OpenReadOnly 로 엶(이관 없음). 초안 도구 미리보기도 OpenReadOnly |
| R3-13 | P3 | 토큰에 비용 모델 지문 · 스키마 · 빌드 포함 | 수용 |
| R3-14 | P3 | 동결 뒤 하네스 미커밋 수정 | 확인 — 3판 초안 변이 하네스(dry-run 20/22, 생존 2: 커밋 전 보고 → 구조 핀 추가 · ADJUSTMENT_CLOSED 필터는 COMPLETED 와 상호배타라 동등 변이) |

### 3차 발견의 4판 처분 (사용자 결정 2026-10-13 3차 — 브로커 공식 1분봉)

| # | 4판 처분 |
|---|---|
| R3-1 | 해소 — 가격 출처를 원장 밖 브로커 공식 1분봉(adjusted=false, 그 조정 시각이 든 봉의 종가)으로. 그 분 봉 없음 → `no_intake_candle`. 체결 트랜잭션 안 네트워크 0(`needs_intake_price`) |
| R3-2 | 해소 — 허용 간격 상수 삭제(그 분의 봉만 받음). 정규장 밖 조정은 실측(23 묶음 중 21 에 그 분 봉, 2 없음)으로 「그 분만」 결정(design D1.4) |
| R3-3 | 수용 — `instanceSellIntegrity`(새 파일): 귀속 주문 확정이 인스턴스 열림 전 → `foreign_instance_sell`, 열린 뒤 확정된 귀속 밖 로컬 매도 → `unattributed_local_sell`. 시각은 `settled_at`(committed_at 은 밀려 적혀 고정 시계 시험에서 헛거절 — 실측), 거절 조건에만. RED 시험 둘(base 에선 조작 행) |
| R3-4 | 수용 — 도구: 읽기 전용 열기 → 버전 = 빌드 → 같을 때만 `Open`. 스키마 32 파일 실행 거절 + 바이트 동일 시험 |
| R3-5 | 수용 — 이관한 사본에서 항목별 쓰기 잠금 실측(최대 22.17 ms), 실행 창 KST 05:00~09:00 를 tasks 게이트 뒤 절차 · spec 에 |
| R3-6 | 수용 — MODIFIED 「성과 원시 지표」: 조건부 동결 · 복구 절차로만 나중 쓰기 · 초기 수량 정의(편입 = 편입 + 묶음, 매도 수량과 같음) |
| R3-7 | 해소 — exit_events 관측을 쓰지 않음. 남은 시각(조정 · 편입 · 주문 확정 · 수명)은 RFC3339Nano 파싱 비교 |
| R3-8 | 해소 — exit 판정 기록 함수들을 근거로 쓰지 않으므로 번들 불요. adoptedBasis 는 무편집(편입 시각은 adoptionEpoch 의 자체 조인) |
| R3-9 | 수용 — D7 시험 · 변이 목록 4판 |
| R3-10 | 수용 — 줄은 best-effort 로 spec · D3 에 명시 |
| R3-11 | 명명 — 쌓인 change 검사는 Manager 판단 |
| R3-12 | 확인 — census · 미리보기 모두 OpenReadOnly |
| R3-13 | 수용 — 토큰에 쓴 분봉 · 비용 모델(`%#v`) · 빌드 스키마 버전 |
| R3-14 | 확인 |

## 4차 proposal-freeze review (2026-10-13, 4판 bc928a29) — APPROVE-WITH-FIXES

- 보이스: 적대적 Eng 1(분리된 컨텍스트, code-reviewer, 읽기 전용 · 네트워크 0 — 저장소 무편집, 스크래치에서 bc928a29 + wip/a130-4 + apply_edits4.py 로 재구성해
  build · vet · 시험 통과 확인). 판정: **APPROVE-WITH-FIXES** — P1 둘은 GREEN 전에, P2 는 구현 task 안에서(계약 동결 불변).
- 성립 확인: R · 체결 트랜잭션 네트워크 0(journal 이 official 을 import 안 함) · 커밋 뒤 보고 · COMPLETED/ADJUSTMENT_CLOSED 배타 · BarLabel · `%#v` 결정성 ·
  BEGIN IMMEDIATE · 가드 순서 · 엔진 싱크 · 번들 15 sha.
- 처분(전부 GREEN 커밋에 반영 — 아래 시험 이름은 커밋된 파일):

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| R4-1 | **P1** | 문서대로(config-dir 없이) 돌리면 도구가 엔진과 **다른 토큰 캐시**를 써 새 토큰 발급 → 엔진 토큰 무효화(a082) | **수용** — 묶음이 있으면 `--config-dir` 필수(client 열기 전 거절) · 토큰 파일 경로 출력 · 런북에 `--config-dir ~/.config/tossctl`. 시험 `TestBarsAreNotReadWithoutTheEngineConfigDir` · `TestFetchPricesKeepsOnlyTheExactMinuteAndPaces`(경로 출력). **사고 기록**: 이 change 의 사본 미리보기 · 실행(05:10~05:11Z)과 `ops call` 실측(04:32Z~)은 기본값 토큰 캐시를 썼고, 05:11:33Z 캐시 토큰 · 05:11:35Z 엔진 config 토큰이 2초 차로 다시 쓰였음 — 엔진이 그 직후 재발급했을 가능성이 큼(장 마감 후, 주문 없음). Manager 보고 |
| R4-2 | **P1** | 「그 분만」 가드가 시험 0 이고 둘째 가드는 요청 이름을 적어 항등 | **수용** — 리더가 돌려받은 이름을 돌려주고 도구가 그것을 적음(priceLots 가 실제 둘째 가드), httptest 시험 `TestA130TheMinuteCandleIsReadUnadjustedAndOnlyForItsOwnMinute`(질의값 · 정확한 분 · 이른 봉 · 빈 봉) · `TestA130ARateLimitedCandleReadFailsAfterOneRequest`(429 1회), 변이 추가 |
| R4-3 | P2 | 앞으로 쪽 flatten 쓸어 담기(닫힌 뒤 확정) 미검사 | **수용** — 닫힌 뒤 확정도 `foreign_instance_sell`, 시험 `TestA130ALaterInstanceFlattenIsNotThisTripsSell`, spec 「열리기 전 또는 닫힌 뒤」 |
| R4-4 | P2 | 체결 트랜잭션 안 추가 시간 미측정 | **수용 · 실측으로 수리** — 판정 소요 사본 측정: READOPTED 를 exit_events 에서 훑던 판정이 최대 20~27 ms(포지션 하나에 27,071 이벤트). exit 상태 `lifecycle_generation > 1` 한 줄 읽기로 바꿔 최대 2.4~2.9 ms(a130 이전 최대 1.7~2.4 ms) — `analysis/verify/receipts.md` |
| R4-5 | P2 | 기존 `StrictMinuteCandles` 를 두고 약한 리더를 새로 만듦 · design 사실 오류 | **수용** — 리더를 `StrictMinuteCandles(count=1, 서울 오프셋 before)` 감싸기로, design 사실 정정 |
| R4-6 | P2 | 규칙 5: `closeExitStateOnAdjustmentTx` · `journalTimeStrictlyAfter` · `sumSellFills` · `Client.send` 번들 없음 | **수용** — 인용 근거 번들 넷 추가 |
| R4-7 | P2 | 실측 하네스 미커밋 | **수용** — `analysis/harness/a130_intake_candles.py`(1분봉 실측) 커밋, census 는 커밋된 시험, 잠금 시간은 도구 출력 |
| R4-8 | P3 | 귀속 밖 매도 창의 시작이 fold-in 시각 | **수용** — 편입 관측 시각으로, 시험 `TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal` |
| R4-9 | P3 | D2 표 `exit_state_unreadable` 은 체결 트랜잭션에서도 나옴 | **수용** — 표 정정 |
| R4-10 | P3 | FLM 의 무결 검사 · sell_unpriced 순서 | **수용** — FLM 편집 뒤 절 · BTM 좌표가 코드 순서(sell_unpriced 먼저) |
| R4-11 | P3 | KST 05:00~09:00 고정 문구는 서머타임에 틀림 | **수용** — 「두 시장이 닫힌 시간」 으로, 창은 운영 절차가 정함 |
| R4-12 | P3 | 체결 트랜잭션 네트워크 0 이 구성으로만 | **수용** — `TestA130TheJournalCannotReachTheNetwork`(go list -deps: net/http · internal/official 없음) |
| R4-13 | P3 | 합성 R 구분 집계에 세 번째 출처 | **수용** — spec · design 에 유도 조건 명시 |

- spec 요구 수정(R4-3 「닫힌 뒤」 · R4-1 config 디렉터리 · 이름 대조 · R4-13)은 이 리뷰의 제안을 그대로 반영한 것이고, 구현 리뷰가 requirement 변경 리뷰를 겸한다.
