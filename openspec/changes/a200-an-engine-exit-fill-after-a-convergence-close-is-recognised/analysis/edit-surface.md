# a200 1.7 · 1.8 — 호출 자리와 편집 표면

CodeGraph **1.6.0**, 이 worktree 에서 새로 색인(파일 3,255 · 노드 56,297 · 엣지 207,790, 원문 `code-context/codegraph-callers.txt`).
CodeGraph 는 훅 값 배선(gateway.go:202 · 204 — 함수 값 대입)을 호출로 못 잡음 → 자리 수는 grep + AST 로 셈.

## 1.7 호출자와 호출 자리(생산)

| 함수 | 생산 호출 자리 | 비고 |
|---|---|---|
| `ProjectPosition` | 배선 1(gateway.go:202) · 실행 1(`runApplyHooks` Project@284) | `runApplyHooks` 생산 호출 2: fills.go:551 RecordFill, strategy_dispatch_runtime.go:1348(BUY 만) |
| `enterReconcileInTx` | 1(ProjectPosition@102) | |
| `ApplyExitFill` | 배선 1(gateway.go:204) · 실행 1(`runApplyHooks` Exit@298) | |
| `freezeTradeOutcomeTx` | 1(ApplyExitFill@1036) | |
| `computeTradeOutcome` | 4(freezeTradeOutcomeTx:130 · BackfillTradeOutcome:898 · backfillPlannedOutcome:165 · tradeOutcomeBackfillPlan:239) | |
| `adoptionEpoch` | 3(adoptedRoundTripLegs:529 · intakeLotsForBackfill:107 · tradeOutcomeBackfillPlan:240) | D5 편입을 살리면 편집 대상 |
| `closeExitStateOnAdjustmentTx` | 1(ApplyPositionAdjustment:356) | 무편집 목표 |
| `currentInstance` | 1(ProjectPosition@78) | (f) |
| `exitContextForFill` | 1(ApplyExitFill@972) | (f) |
| `ExitProposalForIntent`(journal) | 1(execgw/a166_replay_origin.go:47) | 인정 행이 `proposed_intent_id` 를 실으면 영향 |

## 1.8 편집할 기존 함수

| 함수 | 왜 | 고정 장치 적중 | 본문을 바꾸면 깨질 수 있는 시험(이름) |
|---|---|---|---|
| `ProjectPosition`(position_projection.go) | 거절 갈래 B5 1 자리 | a166 census 0 · 정적 허용 목록: `adoptionIDReaders`(adoption_static_test.go:43 — 파일 단위, 이미 등재) | `TestTheProjectionAndTheFillAreOneCommit` · `TestTheProjectionCarriesThePreviousSnapshotForItsCostBasis`(CodeGraph 가 본 직접 호출 시험), position_projection_test.go 의 거절 → RECONCILE 시험군, 이 change 의 재현 시험 2 개(의도된 RED) |
| `ApplyExitFill`(apply_hook.go) | 완결 · 성과 조건 B14 | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`(apply_hook_test.go:567 — 새 파일이 `pending_intent_id` 등 4 열을 언급하면 실패) · a166 census 0 | exit_state_test.go 의 fill 경로 시험군, trade_outcomes_test.go 동결 시험군, `TestA130ARolledBackCloseReportsNothing`, `TestReadoptedGenerationFillAndCompletionEventsStayGenerationTwo` |
| (조건부) `adoptionEpoch` | D5 를 편입 포지션까지 | `TestA130BarPricedRowsAreDerivableFromTheLedger` · `a130Prices` 하네스 · `TestA130JournalCopyCensus`(trade_outcome 사본 census) · 백필 계획 2 자리가 같이 바뀜 | a130_adopted_outcome_test.go 의 `post_adoption_disposal` 시험들 |
| (조건부) `ExitProposalForIntent` | 인정 행이 `proposed_intent_id` 를 실을 때 | `TestA166ExactExitOrigin` · `TestA166OriginRejectsForeignRowsAndBound`(journal/a166_replay_origin_test.go) | 같음 |
| (조건부) `currentInstance` · `exitContextForFill` | (f) 를 고칠 때만 | 투영 전체 | position_projection_test.go 전부 |

새 파일(인정 판정 · 기록)은 기존 함수 수에 안 들어감. 단 판정이 `pending_intent_id` 를 읽으려면 `ApplyTx.PendingState` 를 써야 함(위 구조 시험).
`bindApplyHooks` · gateway 배선은 바꿀 필요 없음(a166 census `gateway.go||buildGateway` · `newGatewayReplay` 는 배선을 바꿀 때만, a130 인자 수 2 단언).
전이표 census(`TestTheTableCoversEveryReachableCombination`) · `ValidReconcileCause` · `TradeOutcomeRefusalReasons` 는 새 값을 안 더하면 무적중.

## `exit_events.action` 소비자와 새 action 의 영향

| 소비자 | 자리 | 새 action 영향 |
|---|---|---|
| exit 출처 판정(a166 예산 우선순위) | journal/a166_replay_origin.go:17-38 → execgw/a166_replay_origin.go:47 | 행에 `proposed_intent_id` 가 있으면 **오분류**(수명 목록 밖 → invalid → 손절 intent 가 Critical 을 잃음). 없으면 무시됨 |
| 보호 청산 시도 존재(a087) | journal/a087_protective_exit_attempt.go:29 `e.action IN (보호 action)` | 무시됨 |
| 성과 매도 귀속(편입) | trade_outcomes.go `engineSellOrders` 598-637 (`proposed_intent_id` 결합, action 무관) | `proposed_intent_id` 를 실으면 같은 intent 중복(DISTINCT 로 흡수) — 무해 |
| 백필 모집단 | trade_outcome_backfill.go:202-210 `action = COMPLETED` | 새 action 은 무시. 단 D5 가 `COMPLETED` 를 쓰면 편입 포지션이 모집단에 들어옴(배타성 주석 :12-16 거짓) |
| 콘솔 · httpapi 이력 | account_views.go:311 · 452, provenance.go:282 | 문자열 그대로 표시(라벨 매핑 없음 — internal/console 에 action 리터럴 0) |
| 평가 증거 검증 | exit_snapshot.go:376-417 | 증거 열을 하나라도 채우면 **깨짐**(그 포지션 `ExitEvents` 가 corrupt) — 수량을 `projected_quantity` 에 실을 수 없음 |
| 일일 손실(a137) | journal/a137_daily_loss_rows.go:137 · dailyloss/measure.go:124 | action 은 안 읽음(조정 행을 읽음). 수렴으로 닫힌 관리 포지션 + 엔진 매도 발송 = 측정 불가 — a200 뒤에도 그대로 |
| 분석 SQL(a087 진행 중) | a087/analysis/exit-slippage-3.8/exit_slippage.sql | action 필터 — 무시됨 |
| 골든 · 시험 데이터 | 저장소 루트 `analysis/` · `internal/*/testdata` 에 `ADJUSTMENT_CLOSED` · `PROPOSAL_FILLED` · `exit_events` 리터럴 0(grep) | 무관 |

## 진행 중 change 와의 겹침

| change | 겹침 |
|---|---|
| a167(보호 잔량 재가격, 착수 전) | proposal 「FLM/BTM 필수」 에 `journal.ApplyExitFill` — **같은 함수 편집 후보**. 착수 조건이 a087 Phase 2 라 지금 활성 편집 없음 |
| a087(보호 청산 시장가) | `ApplyExitFill` · `position_projection.go` 를 인용만(편집 아님) |
| a197(보류) · a171 · a198 | 후보 함수 · 파일 언급 0 |
| a199(인터록, 다른 팀메이트) | 이 1단계에서 읽지 않음(worktree 접근 금지) — Manager 가 대조 |
