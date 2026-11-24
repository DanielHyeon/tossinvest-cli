# a200 1.7 · 1.8 — 호출 자리와 편집 표면 (범위 개정 2026-11-24 판)

CodeGraph **1.6.0**, 이 worktree 에서 새로 색인(파일 3,255 · 노드 56,297 · 엣지 207,790, 원문 `code-context/codegraph-callers.txt`).
CodeGraph 는 훅 값 배선(gateway.go:202 · 204 — 함수 값 대입)을 호출로 못 잡음 → 자리 수는 grep + AST 로 셈.

## 1.7 호출자와 호출 자리(생산)

| 함수 | 생산 호출 자리 | 비고 |
|---|---|---|
| `ProjectPosition` | 배선 1(gateway.go:202) · 실행 1(`runApplyHooks` Project@284) | `runApplyHooks` 생산 호출 2: fills.go:551 RecordFill, strategy_dispatch_runtime.go:1348(BUY 만 — 매도 거절에 안 닿음) |
| `enterReconcileInTx` | 1(ProjectPosition@102) | |
| `ApplyExitFill` | 배선 1(gateway.go:204) · 실행 1(`runApplyHooks` Exit@298) | **무편집** |
| `freezeTradeOutcomeTx` | 1(ApplyExitFill@1036) | 무편집 |
| `computeTradeOutcome` | 4 | 무편집 |
| `adoptionEpoch` | 3 | 무편집(성과 비목표) |
| `closeExitStateOnAdjustmentTx` | 1(ApplyPositionAdjustment:356) | 무편집 |
| `currentInstance` · `exitContextForFill` | 각 1 | 무편집((f) 범위 밖) |
| `appendExitEventFromHook` | ApplyExitFill:1014 · markExitCompletedTx:969 | 표지 쓰기에 **그대로 호출**(무편집) |
| `ApplyTx.PendingState` | ApplyExitFill:982 · MoveTakenRatioTotal:430 | 판정이 그대로 호출(무편집) |

## 1.8 확정 편집 표면

| 함수 | 왜 | 고정 장치 적중 | 본문을 바꾸면 깨질 수 있는 시험 |
|---|---|---|---|
| `ProjectPosition`(position_projection.go) — **유일한 기존 함수** | B5(if@101) 안, enterReconcileInTx@102 직전 판정 1 자리 | a166 census 0 · `adoptionIDReaders`(파일 단위, 이미 등재 — 새 파일은 그 열을 안 읽음) | `TestAnOversellFreezesTheProjectionAndBlocksTheSymbol` · `TestASellWithNothingHeldIsAttributionFailure` · `TestTheProjectionAndTheFillAreOneCommit` · `TestTheProjectionCarriesThePreviousSnapshotForItsCostBasis`, 재현 시험 `…AConvergenceCloseThenTheStopFillIsAnAttributionFailure` 와 cycles 의 W4 · W5(의도된 RED) |
| 새 파일(판정 · 표지) | design D2 · D2a · D3 | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` — 새 파일에 `pending_intent_id` 등 4 열 철자 금지 → `PendingState` 로 읽음. 판정이 읽는 표는 `position_adjustments` · `mutation_attempts` · `intents`(체결 누적 표는 안 읽음 — Manager 결정 3). 그 읽기를 막는 정적 목록 없음(grep) | — |

FLM/BTM pre-edit 번들: `function-logic/internal-journal--projectposition/`(check_analysis: required 0 · 번들 검증 통과).

## 표지 행 — `exit_events` 소비자 전수와 영향

표지 모양: `position_id` · `action`(새 상수) · `created_at` · `lifecycle_generation`(appendExitEventFromHook 이 exit 상태에서 복사)만 채움. `proposed_intent_id` · `decision_id` · 가격 · 레벨 · 평가 증거 열 전부 NULL.
생산 코드의 `exit_events` 읽기 전수(grep `exit_events` — 비시험 Go, internal · cmd · tools)를 행마다 대조.

| 소비자 | 자리 | 영향 | 근거 |
|---|---|---|---|
| exit 출처(a166 예산 우선순위) | journal/a166_replay_origin.go:17-18 · :59 | **무시됨** | `WHERE proposed_intent_id = ?` — NULL 행은 안 잡힘 |
| 보호 청산 시도 존재(a087) | a087_protective_exit_attempt.go:26 | **무시됨** | `JOIN … ON e.proposed_intent_id = i.id` |
| 편입 성과 매도 귀속 | trade_outcomes.go:614 `engineSellOrders` | **무시됨** | `JOIN … a.intent_id = e.proposed_intent_id` + `coalesce(…) <> ''` |
| 백필 모집단 | trade_outcome_backfill.go:209 | **무시됨** | `e.action = 'COMPLETED'` |
| 결정 중복 검사 | exit_state.go:494 | **무시됨** | `WHERE decision_id = ?` — NULL |
| provenance exit intent | provenance.go:240 | **무시됨** | `proposed_intent_id IS NOT NULL AND <> ''` |
| provenance 사건 목록 | provenance.go:284 | **원문 표시만** | 행 하나가 `EXIT_EVENT`(action 문자열)로 나열됨 — 판정 없음 |
| 콘솔 이력(포트폴리오 · 개요) | account_views.go:319 → console/portfolio.go:738 · overview.go:580 | **원문 표시만** | action 문자열 그대로, 가격 칸 「—」. 라벨 매핑 · action 분기 0(internal/console grep) |
| 콘솔 intent 별 행 | account_views.go:450 · 467-468 | **무시됨** | intent 키 |
| 평가 증거 검증 | exit_snapshot.go:376-417(`hydrateExitEventEvidence`) | **무시됨** | 증거 열이 모두 NULL → `any()` 거짓 → legacy 갈래(오류 없음) |
| `Journal.ExitEvents` | exit_state.go:881 | 생산 호출 0 | grep |
| readonly 표 목록 | readonly.go:79-92 | 무관 | 읽기 전용 연결의 표 · 열 목록 |
| 일일 손실(a137) | — | 무관 | exit_events 를 안 읽음(조정 행을 읽음) |
| 분석 SQL(a087 진행 중) | a087/analysis/exit-slippage-3.8/exit_slippage.sql | **무시됨** | intent 결합 + action 필터 |
| 골든 · 시험 데이터 | 루트 `analysis/` · `internal/*/testdata` | 무관 | 리터럴 0(grep) |

**판정: 오분류 0 · 깨짐 0** — 판정에 닿는 소비자는 전부 무시, 화면 둘은 원문을 한 줄 더 보여 줄 뿐 → **표지 행을 쓴다**(Manager 승인 2026-11-24 — 원문 표시는 판정에 쓰이지 않으므로 「무시됨」).

## 진행 중 change 와의 겹침

| change | 겹침 |
|---|---|
| a167(보호 잔량 재가격, 착수 전) | `ApplyExitFill` 을 FLM 대상으로 적음 — a200 은 `ApplyExitFill` **무편집**이라 편집 겹침 없음. 단 a167 이 청산을 정정(AMEND)하면 그 체결은 a200 판정 (라)의 「PLACE · AMEND 하나」 에서 거절된다 — 거절하는 정상 입력 (나)가 넓어짐 |
| a087(보호 청산 시장가) | `ApplyExitFill` · `position_projection.go` 를 인용만(편집 아님). a087 의 `exit_slippage.sql` 은 표지를 무시 |
| a197(보류) · a171 · a198 | 후보 함수 · 파일 언급 0 |
| a199(인터록) | 이 단계에서 읽지 않음(worktree 접근 금지) — Manager 대조 |
