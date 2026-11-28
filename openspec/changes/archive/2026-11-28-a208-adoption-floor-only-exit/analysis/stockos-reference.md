# StockOS 편입 청산 정책 — 원문 재확인

앞선 비교(에이전트 작성, 미검증)를 StockOS 원문으로 다시 읽음. 읽기 전용, StockOS 저장소 무변경.

- 저장소: `/mnt/D/Axipient/workspace/stockos`, 브랜치 `main` = `origin/main` = `fcad869e`(2026-08-17, 로컬 ref 중 가장 최근).
  원격을 새로 받지 않았으므로 그 뒤 StockOS 변경은 **안 봤음**.
- 파일: `apps/api/stockos_api/auto_exit_execution.py`(이하 `aee.py`). Python 이라 `tools/logic-map` 대상이 아님 — 줄 인용만 함.

## 1. 편입 사다리 — 부분익절 제거는 RUNNER 에만

`aee.py:6766-6807` `_resolve_adoption_ladder_policy(exit_policy, *, external_position=True)`:

- `:6796` `if kind is not ExitPolicyKind.COMMON_LADDER_RUNNER or not external_position: return policy`
  — BALANCED · HYBRID 편입은 **등록 사다리 그대로**(부분익절 포함).
- `:6802-6807` RUNNER ∧ 외부 보유만 `ProfitLadderPolicy(policy_id="external_runner_v1", rungs=… partial_ratio=0.0, final_take_full=False)`.
- 시험: `tests/test_auto_exit_execution.py:4065-4066`(`final_take_full is False`, 모든 rung `partial_ratio == 0.0`).
- RUNNER 표: `apps/web/tests/e2e/exit-policy-selection.spec.ts:25-27` — 2.5/0/0 · 4.5/2/0.15 · 7/3.5/0 (TossOS `common_policy.go:48-53` 과 같음).

**정정.** 앞선 비교의 「StockOS 는 편입 포지션에 부분익절 없는 자기 사다리를 준다」는 **RUNNER 일 때만** 참. TossOS 는 같은 규칙을
이미 가짐(`internal/exitpolicy/common_policy.go:104-127` 주석이 StockOS A168 을 인용). 사용자가 승인한 「부분익절 없음」을
모든 편입에 적용하는 것은 StockOS 보다 **한 걸음 더 나간** 결정이고, design D1 이 그 결정을 RUNNER 표 재사용으로 구현함.

## 2. 편입 직후 60초 유예(settle window)

- `aee.py:287-293` 상수 `ADOPTION_SETTLE_WINDOW_SECONDS = 60.0`. 주석(287-292): 편입 포지션의 첫 시세가 브로커 스냅숏 가격이라
  「잠재 평가이익을 첫 익절 목표 도달로 해석해 즉시 청산」한 사고(TAKE_PARTIAL_PROFIT 이 state backfill 3ms 뒤 발화).
- `aee.py:295-308` `_PROFIT_SIDE_ACTIONS` — TAKE_PARTIAL_PROFIT · TAKE_FULL_PROFIT · TRAIL_PROFIT · RAISE_PROFIT_LOCK · LADDER_TAKE_PROFIT ·
  LADDER_PARTIAL · **MOVE_STOP_TO_BREAKEVEN**(`:307`, 주석 302-306: 유예 중 floor 를 기록하면 만료 뒤 소폭 조정에 floor 손절이 발화).
- `aee.py:6978-6998` `_is_in_adoption_settle_window` — 기준은 `strategy_profile_id == "manual-adoption"` 인 상태 행의
  `created_at`(`:6992`, 모델 기본값 `_utcnow` — `packages/trading/stockos_trading/persistence/models.py:75`) 과 `now`(벽시계 UTC)의 차,
  `:6998` `age < 60`. 영속된 시각이라 재시작에도 같은 창.
- `aee.py:7001-7024` `_apply_adoption_settle_window` — 창 안이고 `_PROFIT_SIDE_ACTIONS` 이면 HOLD 로 낮추고 사유
  `ADOPTION_SETTLE_WINDOW:<원 action>` 을 남김(`:7016` 판정). docstring: hard stop · time exit 는 영향 없음.
- `aee.py:6868-6869` `_resolve_auto_exit_ladder_policy` — 창 안이면 사다리 자체를 None 으로(rung 이 묵은 시세로 영구 승격되는 것을 막음,
  주석 `:1604-1610`).
- `aee.py:1630-1639` 같은 `now_ts` 로 두 창 판정을 묶음(경계를 가로지르지 않게).
- `aee.py:2483-2492` `_update_high_since_entry` — 창 안에서는 고점 후보에 현재 시세를 넣지 않음(**고점 동결**).

## 3. 편입 손절 기준 — 평균단가(알려진 결함)

`aee.py:1860-1876` `_build_adopted_position_state` — `entry_price = record.avg_price`(`:1867`),
`initial_stop = entry_price × (1 + hard_stop_loss_pct/100)`(`:1871`, 인자 기본 −1.2). 현재가가 이미 그 선 아래인 보유를 편입하면
첫 관측에서 손절이 남 — 사용자가 가져오지 말라고 한 결함. TossOS 는 편입 관측가 × (1 − pct), pct ≥ 0.02
(`internal/exitpolicy/adoption.go:46-72`, `internal/app/engine/adoption.go:374`) 를 유지함.

## 4. TossOS 로 옮길 때 달라지는 전제

StockOS 유예의 원인(평균단가 기준 + 브로커 스냅숏 시세 → 편입 순간 이미 「이익」) 은 TossOS 에 없음: TossOS 의 기준선 · 고점은 편입 관측가
자체로 열리므로(`internal/journal/adoption.go:266-290` 주석, R = 0) 편입 순간의 평가이익은 0. TossOS 에서 유예가 막는 것은
「편입 관측(최대 15초 묵음, exit-policy 스펙) 과 첫 exit 관측 사이의 가격 차이로 첫 rung 이 바로 승격되는 것」 뿐 — design 4 · 사람 결정 Q1.
