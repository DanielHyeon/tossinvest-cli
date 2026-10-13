# a130 design (3판)

base `233d5d9c`. 분기 주장은 전부 `analysis/function-logic/*/ast.json`(편집 전 추출) 의 열거에서 온다.

판 이력: 1판 REJECT(review.md F-1~F-13) → 2판(사용자 결정 (b) 「거절만」, 2차 리뷰 APPROVE-WITH-FIXES R2-1~R2-11) →
**3판: 사용자가 (b) 를 철회(2026-10-13 10:4x) — 재계산한다.** 늘어난 묶음의 매입가 = 그 조정 뒤 엔진이 관측한 시장가(묶음별 편입가).
편입 시점 관측가를 늘어난 묶음에도 대입하는 안(a)은 기각(늦게 늘어난 종목에 틀린 손익 박제), converge 의 평균가 기록은 a132(별도).

## 사실 (편집 전 AST · HEAD · 저널 사본 실측)

- `freezeTradeOutcomeTx`(trade_outcomes.go:129, 분기 1)의 유일한 호출자는 `ApplyExitFill` B14(:1013) 안 :1023 이고 반환 bool 을 버린다.
- `computeTradeOutcome`(:163, 분기 12 · return 10): 거절 return 9개(:169 · :181 · :185 · :190 · :197 · :202 · :212 · :218 · :222).
  B5 :197(결정 · 편입 둘 다 없음)은 base 커버리지 0 — 외부 포지션은 exit 상태가 없어 `ApplyExitFill` B3 에서 먼저 돌아간다.
- `adoptedRoundTripLegs`(:488, 분기 5): 매수 다리 = 편입 수량 × 편입 관측가, B3(:504) 엔진 매도 주문 0, B5(:512) 매도 수량 ≠ **편입 시점 수량** → :516.
- 늘어난 수량의 매입가: 대사 수렴 `Converger.ConvergeQuantities`(reconcile/converge.go:142, 요청 :209-226)는 `NewAvgPrice: ""`(:221)를 보내고
  `Journal.ApplyPositionAdjustment`(position_adjustments.go:202)는 `firstNonEmpty(req.NewAvgPrice, target.AvgPrice)`(:312)를 저장한다 — 조정 행의 평균가는
  편입 전 원 보유분의 cost_basis 를 이어 적은 값이고 매입가가 아니다(리뷰 F-1).
- 엔진의 가격 관측은 `exit_events` 의 `observed_price` 와 `observed_at`(a042 스냅숏 판정, RFC3339Nano; 없으면 a042 이전 기록 — 기록 시각 `created_at`)이다.
  관측 루프 주기 `DefaultExitObservationInterval` = 5초(exitloop.go:98).
- 재편입 `resetExitStateForReadoptTx`(apply_hook.go:699)는 exit 상태의 `entry_price` · `initial_risk` 를 다시 쓰고 `ExitEventReadopted` 를 남긴다(:739-743).
  유일한 호출자는 `ApplyPositionPolicy`(position_policy.go:145).
- `engineSellOrders`(:551)의 flatten 갈래(:577-591)는 계좌 · 심볼 · 시장으로만 묶여 인스턴스 · 시각 경계가 없다(F-3).
- `Journal.RecordFill`(fills.go:313, 분기 38): 훅 B37(:548), 커밋 B38(:553). `runApplyHooks` 의 다른 생산 호출자
  `backfillConfirmedStrategyFillTx`(strategy_dispatch_runtime.go:1297)는 `Side: "BUY"` 만 적용한다(:1303 · :1316) — 매수는 포지션을 닫지 못한다.
- `markExitCompletedTx`(exit_state.go:934)만 `COMPLETED` exit 이벤트를 쓰고 그것은 `ApplyExitFill` B14(체결로 닫힘) 안이다. 대사로 닫힌 편입 포지션은
  `ADJUSTMENT_CLOSED`(position_adjustments.go:427).

## D1. 편입 포지션의 매수 다리 — 묶음별 편입가 (사용자 결정 2026-10-13 2차)

```
q_a   = position_adoptions.quantity           편입 수량
p_a   = position_adoptions.observed_price     편입 관측가
t_a   = position_adoptions.observed_at        편입 관측 시각
r     = exit_states.initial_risk              주당 최초 위험 — 편입 시점 값, 불변(= p_a − synthetic_stop)

lot k = position_adjustments 행 중 position_id = 이 포지션, created_at_k > t_a, new_k > prev_k
Δ⁺_k  = new_k − prev_k
p_k   = 이 포지션의 exit_events 가격 관측 중 관측 시각 ≥ created_at_k 인 첫 관측의 observed_price,
        단 (그 관측 시각 − created_at_k) ≤ T_intake(= 383초). 없으면 거절(no_intake_observation)

Q_buy = q_a + Σ_k Δ⁺_k
N_buy = q_a·p_a + Σ_k Δ⁺_k·p_k
동결 조건: 재편입 없음 ∧ 편입 뒤 감소 없음 ∧ 모든 lot 에 p_k 있음 ∧ sell.priced ∧ Q_sell = Q_buy
pnl   = N_sell − N_buy − cost_buy(N_buy) − cost_sell(N_sell)
R     = pnl ÷ (r × Q_sell)          분자 · 분모가 같은 Q_sell(= Q_buy) — 분모를 q_a 로 두면 R 이 부풂
initial_quantity 칼럼 = Q_buy
```

- **A7 과의 정합**: 편입 규칙(A7 round 3: 기준선은 매입 원가가 아니라 편입 시점 관측가, cost_basis 는 산식에 안 들어감)을 묶음마다 같은
  규칙으로 적용한다. 조정 행의 평균가 · cost_basis 는 쓰지 않는다.
- **T_intake = 383초의 근거(실측)**: 운영 저널 사본(2026-10-12T21:34Z, 스키마 32)에서 같은 포지션의 연속한 가격 관측(57 포지션, 17,043 간격)의
  간격 분포 — p50 5.369초 · p90 5.735초 · p95 10.144초 · p98 54.645초 · **p99 382.777초** · p99.5 1,724.673초. 5초 관측 주기대로면 대부분 6초 안이고,
  p99 를 넘는 간격은 장 마감 · 휴장 · 엔진 정지를 건너뛴 것이다. 그 너머의 관측은 그 묶음이 생긴 순간의 시장가가 아닐 가능성이 크므로 받지 않는다.
  p99 를 초 단위로 올린 383초를 상수로 둔다(`intakeObservationTolerance`). 영수증: `analysis/verify/intake-tolerance.md`.
- **「조정 뒤 첫 관측」 인 이유**: 조정은 그 시각의 계좌 스냅숏(broker_as_of)을 프로젝션에 반영한 것이고, 엔진은 그 뒤 관측부터 늘어난 수량을
  보호한다(사용자 결정이 정한 방향). 조정 앞 관측은 쓰지 않는다.
- **편입 뒤 감소(F-3)**: `created_at ≥ t_a` 인 감소가 있으면 거절(`post_adoption_disposal`). flatten 쓸어 담기(F)와 사람 처분(H)이 맞아떨어져
  Q_sell = Q_buy 가 되는 조작 행을 막는다. **이름 붙인 회귀(F-9)**: 편입 3 → +3 → −3 → 엔진 3주 매도는 a130 이전엔 동결(매도 = 편입 수량),
  a130 뒤엔 `post_adoption_disposal`. 사본에서 동결된 31행 중 이 규칙에 걸리는 것 0(census).
- **재편입(I-2)**: `ExitEventReadopted` 가 있으면 거절(`readopted`) — r 이 다시 쓰였다.
- **같은 초 경계**: 증가는 엄격(`>`, 같은 초의 증가 · fold-in(0→q)은 편입 수량에 들어 있음), 감소는 포함(`>=`, 순서를 못 가르는 감소는 안전 쪽).
  어느 쪽으로 틀려도 수량 불일치 · 거절로 끝나고 조작된 행은 생기지 않는다.
- **귀속**: 조정 · 관측 모두 `position_id` 명시 참조. 시각은 같은 인스턴스의 사건 순서와 관측 지연만 가른다(시간창 귀속 금지 계약과 무관).
- 편입 뒤 증가가 없으면 Q_buy = q_a, N_buy = q_a·p_a — **a130 이전 산식과 같다**(사본 동결 31행 재계산 31/31 일치).

## D2. 사유 코드 — 판정 함수가 돌려주는 타입

`computeTradeOutcome` · `adoptedRoundTripLegs` 는 `TradeOutcomeRefusalReason`(새 파일 `trade_outcome_refusal.go` 의 string 타입, 빈 값 = 성공)을 돌려준다.
값은 그 파일의 상수 17개가 전부다(시험이 go/parser 로 상수 선언과 목록을 대조). 사유는 문구에서 유도하지 않는다.

| 코드 | 자리 | 런타임 줄에 나올 수 있나 |
|---|---|---|
| `cost_model_unconfigured` | computeTradeOutcome :169 | 예(비용 모델 미구성 배선) |
| `position_unreadable` | :181 · :185 · :190 | 예(읽기 실패 — 일시) |
| `unmanaged_position` | :197 | 아니오 — exit 상태가 없어 ApplyExitFill B3 에서 먼저 돌아감(census · backfill 만) |
| `exit_state_unreadable` | :202 | 아니오 — 같은 이유(census · backfill 만) |
| `round_trip_unpriceable` | :212 결정 경로 `roundTripLegs` false | 예 — 진입 체결 없음 · 미관측 평균가 · 읽기 실패(일시)를 묶음 |
| `empty_buy_leg` | :212 `buy.quantity.Sign() <= 0` | 예 |
| `cost_estimate_failed` | :218 · :222 | 예 |
| `adoption_unreadable` | adoptedRoundTripLegs B1 | 예 — 기록 없음 · 해석 실패를 묶음 |
| `readopted` · `post_adoption_disposal` · `no_intake_observation` · `adoption_epoch_unreadable` | D1 편입 시기 | 예 |
| `engine_sell_unreadable` | B2 · B4 | 예(일시) |
| `no_engine_sell` | B3 | 예 |
| `sell_unpriced` · `sell_quantity_mismatch` | B5 | 예 |
| `insert_failed` | freezeTradeOutcomeTx INSERT 오류 | 예 |

`freezeTradeOutcomeTx` 는 사유를 돌려준다. `Journal.BackfillTradeOutcome` 의 거절 오류 문장에 사유 코드를 덧붙인다.

## D3. 거절 기록 — 커밋이 성공한 뒤에만, 계좌 없이 (F-8)

- `ApplyHooks.OutcomeRefused func(TradeOutcomeRefusal)`. `TradeOutcomeRefusal` = {Symbol, Market, InstanceSeq, Reason} — 계좌 참조 · 포지션 id 없음.
- `ApplyExitFill` B14: 사유가 비어 있지 않으면 `tx.noteOutcomeRefusal(ctx, positionID, reason)` — 싱크가 묶였고 ctx 에 수집 목록이 있을 때만 같은
  트랜잭션에서 `positions` 한 줄(market · symbol · instance_seq)을 읽어 그 목록에 쌓기만 한다. 오류를 돌려주지 않는다.
- 수집 목록은 `RecordFill` 이 만든다: B37 의 `runApplyHooks` 호출에 `withOutcomeRefusals(ctx, &refusals)` 를 넘기고, `tx.Commit()` 성공(B38 다음)에만
  `j.reportOutcomeRefusals(refusals)` — 싱크 호출마다 `recover`. 롤백된 종결은 알리지 않는다. 롤백 뒤 같은 체결이 재적용되어 커밋되면 그때 한 번.
- `runApplyHooks` 시그니처는 그대로(시험 호출자 3곳). 바뀌는 것은 핸들 리터럴의 `outcomeRefused` 필드 하나. 전략 되메움 경로는 무편집 —
  매수만 적용하므로 거절이 생기지 않는다. 목록 없이 싱크만 있는 미래 호출자는 거절을 조용히 버린다 — D5 잔여.
- 엔진 싱크(`internal/app/engine/trade_outcome_refusal_log.go`)는 줄 쓰기를 새 goroutine 으로 넘기고 곧바로 돌아온다(막힌 로그 대상이 체결 감지 루프를
  붙잡지 않게). goroutine 안의 패닉도 거둔다. 줄: `Warn("trade_outcome.not_frozen", "symbol", …, "market", …, "instance_seq", …, "reason", …)`.
- 배선: `NewContext` 가 `bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger))`(생산 `opts.Logger` = engine.log 로거). `bindApplyHooks` 는 선택 인자
  (가변 인자 하나)를 받아 시험 호출자 14곳이 그대로 컴파일되고 싱크 없이 a130 이전과 같다. 리터럴은 gateway.go 에 남는다(campaign 정적 시험).
- 편입 시기 읽기 · backfill 모집단은 `positions.adoption_id` 를 쓴다 — `adoption_static_test.go` 읽기 허용 목록(패키지 변수, 함수 아님)에
  `trade_outcome_refusal.go` · `trade_outcome_backfill.go` 를 더한다(둘 다 UPDATE 없음).
- **싱크가 nil 이면** 거절 경로는 추가 읽기 · 호출이 없다. 편입 시기 · 관측 읽기는 편입 종결마다 싱크와 무관하게 돈다(같은 트랜잭션 안의 읽기 셋).

## D4. backfill — 한정 모집단 · 미리보기 → 사람 확인 → 항목별 재대조 실행 (F-2 · F-7)

- **모집단**: `state = 'CLOSED'` ∧ 편입 기록 있음 ∧ `COMPLETED` exit 이벤트 있음(체결로 닫힘) ∧ `ADJUSTMENT_CLOSED` 없음 ∧ 성과 행 없음. 보존 기간(180일)
  밖에서 종결된 항목은 `outside_retention` 으로 표시만 하고 채우지 않는다(보존 정리가 지운 행을 되살리지 않음). 엔진 진입 포지션은 넣지 않는다 —
  `roundTripLegs` 는 매도 수량을 검사하지 않고 체결 창에 상한이 없어 닫는 트랜잭션 밖에서는 정확하지 않다(F-2).
- **미리보기** `PreviewTradeOutcomeBackfill(ctx, *ReadOnly, model, now)` — 읽기 전용 연결(mode=ro + query_only)로 항목마다 판정(fill · refused ·
  outside_retention) · 계산 결과 · 묶음별 편입가 관측 지연 · 거절 사유와 확인 토큰(채움 대상의 정규 직렬화 SHA-256 앞 16 hex). 쓰기 0
  (`-shm` 의 읽기 표시는 operator-console spec 의 읽기 전용 열기 예외).
- **실행** `(*Journal).BackfillTradeOutcomes(ctx, model, now, token)` — 계획을 다시 계산해 토큰이 다르면 0 행. 같으면 채움 대상마다 **짧은 BEGIN
  IMMEDIATE 하나**에서 행 존재 확인 → 다시 계산 → 계획 값과 모든 칸 대조(다르면 `ErrBackfillItemMoved`, 그 항목과 뒤는 쓰지 않음) → INSERT → 커밋.
  쓰기 잠금은 항목 하나 동안만(손절 기록을 오래 막지 않음, 불변식 4). 계획 자체는 트랜잭션 없이 읽는다 — 정합은 항목별 재대조가 보장한다.
- **도구** `tools/a130-outcome-backfill`: 기본 = 미리보기, `--execute --confirm <토큰>` = 실행. `--journal` 경로 필수(기본값 없음). 실행은 원장 스키마가
  이 빌드와 **같을 때만** 쓰기 연결을 연다 — 도구는 원장을 이관하지 않는다(운영 저널은 스키마 32, HEAD 35 — 이관은 엔진 배포의 몫). 비용 모델은
  엔진과 같은 `costs.DefaultModel()`. 출력에 계좌 참조 · 포지션 id 없음. ⚠ mutating — 에이전트 · 시험은 운영 저널에 실행하지 않는다.
- **사본 측정**(`analysis/verify/census-journal-copy.md`, 수 · 사유만): 모집단 12 → fill 3 · `no_intake_observation` 8 · `sell_unpriced` 1. 동결 31행
  재계산 31/31 일치, a130 판정이 이제 거절하는 동결 행 0.

## D5. 잔여 (명명)

- 묶음별 편입가는 「조정 뒤 첫 엔진 관측」 이다 — 실제 매입가가 아니다(실제 매입가는 a132 이후 converge 가 기록하기 전엔 원장에 없다). 383초 안
  관측이 없는 묶음은 행이 없다(사본 8건).
- 관측 시각이 없는 a042 이전 판정 기록은 기록 시각으로 센다(판정은 관측 직후 기록).
- `AdoptPosition`(adoption.go:179)은 요청 수량이 그 순간 인스턴스 수량과 같은지 확인하지 않는다. 감소 규칙의 시각 경계는 fold-in · 대사 수렴 · 편입이
  한 goroutine 에서 차례로 도는 것(reconcileloop)에 기댄다. 시계 어긋남으로 편입 전 증가가 편입 뒤로 읽히면 묶음으로 잡혀 Q_buy > Q_sell → 불일치 거절.
- 사유 코드 몇은 일시 실패와 구조 거절을 묶는다(`round_trip_unpriceable` · `adoption_unreadable` · `*_unreadable`) — 운영자는 일시 실패를 영구로 읽지 말 것.
- `roundTripLegs`(결정 경로)의 트랜잭션 밖 부정확(F-2)과 생산 호출자 0 인 `BackfillTradeOutcome` 의 그 위험은 그대로. a130 의 backfill 은 그 경로를 쓰지 않는다.
- 엔진 줄은 goroutine 이라 프로세스가 곧바로 끝나면 유실될 수 있다(원장은 정본). 목록 없이 싱크만 있는 `runApplyHooks` 호출자는 거절을 버린다(D3).

## D6. 콘솔 안내문 (F-13)

`templates_portfolio.go` 「이 표의 한계」: 성과 행은 엔진이 관리하던 포지션이 엔진 매도로 닫히고 수량 · 가격이 맞을 때만 동결된다는 것, 행이 없는 실제 사유
(엔진이 관리하지 않던 포지션 · 외부 매도 · 대사 조정으로 닫힘 · 편입 뒤 늘어난 수량의 그 시점 관측가가 없음 · 편입 뒤 수량이 줄었거나 재편입됨 ·
매도 수량 불일치 · 평균가 미관측), 거절 사유가 `engine.log` 의 `trade_outcome.not_frozen` 줄에 남는다는 것. 고정 문장 「외부 매도로 종결된 포지션은
여기에 행이 생기지 않으므로」 는 그대로. 바꾼 화면은 실제로 띄워 확인한다(규칙 13).

## D7. 시험 (RED 먼저) · 변이

시험(`internal/journal/a130_adopted_outcome_test.go` · `a130_trade_outcome_backfill_test.go` · `a130_journal_copy_census_test.go`(env 로 켬) ·
`internal/app/engine/a130_trade_outcome_refusal_log_test.go` · `internal/console/a130_history_note_test.go` · `tools/a130-outcome-backfill/main_test.go`):

1. `TestA130AQuantityRaisedAfterAdoptionFreezesWithPerLotIntake` — 편입 3 → 63초 뒤 +3 → 10초 뒤 관측 69000 → 엔진 6주 매도: 행, 수량 6, N_buy =
   3·70000 + 3·69000, R = pnl ÷ (3500 × 6) 정확값, 2배 값 아님, 조정 평균가로 매긴 값 아님, 거절 0.
2. `TestA130TheIntakeToleranceIsExactly383Seconds`(383초 받음 · 384초 거절) · `TestA130AnUnobservedLotIsRefusedWithItsReason`(사유 · 필드 · 계좌 없음).
3. `TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal`(F-9) · `TestA130APostAdoptionDisposalBlocksASweptFlatten`(F-3 — base 에선 조작 행) ·
   `TestA130AReadoptedPositionIsRefused`(I-2) · `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAMismatch` · `TestA130AnUnpricedEngineSellIsReportedAsSuch`.
4. `TestA130APreAdoptionAdjustmentStillFreezes` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot`(N초 앞 · 같은 초) · `TestA130TheSameSecondBoundaryFailsClosed` ·
   `TestA130AnEngineEnteredRoundTripIsUnchanged`.
5. `TestA130ARolledBackCloseReportsNothing`(커밋 실패 → 0, 재적용 커밋 → 1) · `TestA130APanickingRefusalSinkDoesNotRollBackTheClose` ·
   `TestA130AnUnconfiguredModelIsReportedWithItsReason`.
6. `TestA130EachReachableRefusalSiteNamesItsOwnReason`(도달 가능한 자리마다 사유 대조) · `TestA130TheReasonSetIsClosedAndTokenShaped` ·
   `TestA130TheRefusalRecordHasOnlyTheWhitelistedFields`.
7. backfill: 모집단 = 편입 · 엔진 종결 · 행 없음만(엔진 진입 · 대사 종결 · 동결 제외), 미리보기 쓰기 0(db · -wal 바이트 · 수정 시각), 토큰 불일치 0 행,
   일치 → 미리 본 행 그대로, 재사용 0 행, 항목 이동 → 안 씀, 보존 밖 안 채움. 도구: 경로 · 확인 없이 거절, 스키마 불일치 실행 거절.
8. 엔진: nil 로거 → nil 싱크, 줄 키 = 공통 + {symbol, market, instance_seq, reason}, 막힌 writer 가 호출을 붙잡지 않음, NewContext 구조 핀.
9. 콘솔: 안내문의 사유 문구 · 고정 문장.

변이(사본 · 무변이 대조군 GREEN 먼저, `analysis/harness/a130_mutate.py`): 증가 묶음 검사 삭제 · 관측 허용 간격 넘김 · 허용 간격 경계(`>` → `>=`) ·
조정 앞 관측 허용 · 증가 엄격 `>` → `>=` · 감소 검사 삭제 · 감소 `>=` → `>` · 재편입 검사 삭제 · 매도 수량 비교를 편입 수량으로 원복 · 묶음 대금 누락 ·
분모 수량 바꿔치기 · 거절 기록 삭제 · 커밋 전 보고 · 싱크 recover 삭제 · 엔진 goroutine 제거 · NewContext 싱크 인자 삭제 · 핸들 싱크 필드 삭제 ·
backfill 모집단 필터 삭제 · 항목 재대조 삭제 · sell_unpriced 를 mismatch 로.
