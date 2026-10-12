# a130 design

> **상태(2026-10-13): proposal-freeze REJECT — 재설계 대기.** D1 의 늘어난 수량 매입가 전제가 사실과 다름(issues.md I-1, 사람 결정 필요),
> D4 backfill 모집단 · D3 줄 시점 등 수용된 발견 F-1~F-13 은 review.md. 이 판은 1판 기록으로 남김.

base `233d5d9c`. 분기 주장은 전부 `analysis/function-logic/*/ast.json`(편집 전 추출) 의 열거에서 온다.

## 사실 (편집 전, AST · 저널 사본 실측)

- `freezeTradeOutcomeTx`(trade_outcomes.go:129–145, 분기 1): `computeTradeOutcome` 이 false 면 B1 → `return false`(:132),
  INSERT 오류면 `return err == nil`(:144) 로 false. 호출자는 `ApplyExitFill` B14(:1013) 안의 :1023 **하나**이고 반환값을 버린다.
- `computeTradeOutcome`(:163–250, 분기 12, return 10): 거절 return 9개(:169 모델 미구성 · :181 · :185 · :190 포지션 읽기 ·
  :197 결정·편입 둘 다 없음 · :202 exit 상태 없음 · :212 다리 거절 또는 매수 수량 ≤ 0 · :218 · :222 비용 추정 실패), 성공 :238.
  B7(:206) 이 편입 경로(`adoptedRoundTripLegs`)와 결정 경로(`roundTripLegs`)를 가른다.
- `adoptedRoundTripLegs`(:488–519, 분기 5): B1 편입 기록 읽기 실패 · B2 엔진 매도 주문 읽기 실패 · B3 엔진 매도 주문 0 ·
  B4 매도 체결 합산 실패 · **B5(:512) `!sell.priced || sell.quantity.Cmp(adopted) != 0`** → `return …, false`(:516).
  `adopted` 는 `adoptedBasis` 가 읽은 `position_adoptions.quantity` — 편입 시점 수량이다.
- 운영 저널 사본(`2026-10-12`, 읽기 전용): CLOSED 73 · 성과 행 31. 행 없는 42 중 B5 수량 불일치 11(전부 편입 뒤 UNKNOWN 조정으로
  수량 증가, 조정의 `prev_avg_price = new_avg_price`) · B5 미관측 평균가 1 · B3(엔진 매도 0) 다수 · :197(비관리) 다수.
  편입 **전**의 조정이 편입 수량에 이미 들어간 경우(조정으로 수량이 는 뒤 편입)도 있다 — 조정 전부를 더하면
  안 되고 편입 뒤의 것만 더해야 한다.

## D1. 편입 포지션의 실현 R (사용자 결정 2026-10-12 1항)

```
q_a   = position_adoptions.quantity              편입 수량
p_a   = position_adoptions.observed_price        편입 관측가(t0)
t_a   = position_adoptions.observed_at           편입 관측 시각
r     = exit_states.initial_risk                 주당 최초 위험 — 편입 시점 값, 불변 (= p_a − synthetic_stop)

A     = { k ∈ position_adjustments : k.position_id = 포지션
                                    ∧ created_at_k > t_a          (편입 뒤)
                                    ∧ new_quantity_k > prev_quantity_k }   (양(+)의 조정)
Δ_k   = new_quantity_k − prev_quantity_k
p_k   = new_avg_price_k                          그 조정의 브로커 평균가

Q_buy = q_a + Σ_{k∈A} Δ_k
N_buy = q_a·p_a + Σ_{k∈A} Δ_k·p_k
Q_sell, N_sell = 엔진 매도 체결의 수량 합 · 한계 대금 합 (engineSellOrders · sumSellFills — 무변경)

동결 조건:  sell.priced ∧ Q_sell = Q_buy           (아니면 행 없음 + 사유)
pnl   = N_sell − N_buy − cost_buy(N_buy) − cost_sell(N_sell)
R     = pnl ÷ (r × Q_sell)                         (분자 · 분모가 같은 수량 Q_sell = Q_buy)
initial_quantity 칼럼 = Q_sell
```

- **분모를 편입 수량으로 두면 안 되는 이유**: 010170 은 q_a = 3, Q_sell = 6 — 분자는 6주 손익인데 분모가 3주 위험이면 R 이 2배로
  부푼다. 분모와 분자를 같은 Q_sell 로 둔다. 주당 위험 r 은 편입 시점 값 그대로다(exit state 의 진입가 · 최초 위험은 엔진 설계상
  불변이고, 손절 판정도 그 값으로 했다).
- **늘어난 수량의 매입가**: 사용자 결정대로 조정의 `new_avg_price`. 실측 11건은 전부 `prev_avg_price = new_avg_price` 라
  `Δ·new_avg` 와 한계값 `new·new_avg − prev·prev_avg` 가 같다. 둘이 다른 경우(사람이 다른 값에 추가 매수)에는 이 정의가 평균가 쪽으로
  당겨진 매입가를 쓴다 — 잔여로 명명(D6). `new_avg_price = ''`(브로커 미보고) 면 매입가가 없으므로 거절(`adjustment_unpriced`).
- **「편입 뒤」를 시각으로 가르는 이유와 그 실패 방향**: 조정은 `position_id` 로 **명시 참조**되어 이 인스턴스에 귀속된다
  (시간창 귀속 금지 계약은 귀속에 관한 것이고, 여기서 시각은 같은 인스턴스의 두 사건 — 편입과 조정 — 의 순서만 가른다).
  둘 다 초 단위 RFC3339 다. 같은 초의 편입 뒤 조정은 빠져 Q_buy < Q_sell, 같은 초의 편입 전 조정은 엄격 비교로 빠진다
  (편입 수량에 이미 들어 있음). 어느 쪽으로 틀려도 수량이 어긋나 **행 없음 + 사유**로 끝난다. 시각을 읽지 못하면 거절
  (`adjustment_unreadable`).
- **교차 경우(round-5)는 그대로 거절**: 음(−)의 조정(사람의 처분)은 Q_buy 에 더하지 않으므로, 엔진이 일부만 팔면
  Q_sell = q_a + ΣΔ⁺ − (사람 처분) ≠ Q_buy → 행 없음. 엔진 진입 포지션(결정 경로)은 이 change 가 만지지 않는다.
- 편입 뒤 양의 조정이 없으면 A = ∅ 이고 Q_buy = q_a, N_buy = q_a·p_a — **편입 이전 계산과 바이트 단위로 같다**(기존 31행 불변).

## D2. 거절 사유 (조용한 거절 제거, 2항)

`computeTradeOutcome` · `adoptedRoundTripLegs` 는 bool 대신 **사유 코드**(빈 문자열 = 성공)를 돌려준다. 코드는 새 파일
`trade_outcome_refusal.go` 의 상수 집합 하나에만 있다 — 고정 토큰(영소문자 · 밑줄)이고 오류 원문 · 계좌 값을 싣지 않는다.

| 코드 | 자리 |
|---|---|
| `cost_model_unconfigured` | computeTradeOutcome :169 |
| `position_unreadable` | :181 · :185 · :190 |
| `unmanaged_position` | :197 |
| `exit_state_unreadable` | :202 |
| `round_trip_unpriceable` | 결정 경로 `roundTripLegs` false(:212) — 내부 갈래는 이 change 밖 |
| `adoption_unreadable` | adoptedRoundTripLegs B1 |
| `engine_sell_unreadable` | B2 · B4 |
| `no_engine_sell` | B3 |
| `adjustment_unreadable` · `adjustment_unpriced` | 새 편입 뒤 조정 읽기 |
| `sell_unpriced` · `sell_quantity_mismatch` | B5 를 둘로 가름 |
| `empty_buy_leg` | :212 의 `buy.quantity.Sign() <= 0` |
| `cost_estimate_failed` | :218 · :222 |
| `insert_failed` | freezeTradeOutcomeTx 의 INSERT 오류 |

`freezeTradeOutcomeTx` 는 사유를 돌려준다. `ApplyExitFill` 은 사유가 비어 있지 않으면 `ApplyTx` 의 싱크로 알린다.
`BackfillTradeOutcome` 의 거절 오류 문장에 사유 코드를 덧붙인다(값은 그대로 `ErrInvalidRequest` 감쌈).

## D3. 거절 기록의 경로 — 로그 줄, 체결 트랜잭션 밖에서

- 원장 이벤트(새 테이블)는 스키마 버전을 올리고 롤백을 백업 복원으로 만든다. `alert_outbox` 는 미전달 행이 신규 진입을 막는
  안전 장치라 분석 거절을 넣으면 매매를 바꾼다. 그래서 **engine.log 줄**로 남긴다.
- 저널은 로거를 갖지 않는다. `ApplyHooks` 에 `OutcomeRefused func(TradeOutcomeRefusal)` 를 더하고(Costs 와 같은 주입),
  `runApplyHooks` 가 그것을 `ApplyTx` 에 싣는다. nil 이면 아무것도 읽지도 부르지도 않는다(이 change 이전과 동일).
- `TradeOutcomeRefusal` = {Symbol, Market, InstanceSeq, Reason}. **계좌 참조 · 포지션 id 없음**(포지션 id 는 계좌에서 유도된
  해시라 a129 R-3 의 잔여와 같은 이유로 뺌). 종목 · 시장 · 회차는 같은 트랜잭션에서 `positions` 한 줄을 읽어 채운다.
- 싱크 호출은 `recover` 로 감싼다 — 싱크의 패닉이 체결 트랜잭션을 되돌리지 못한다.
- 엔진 싱크(`internal/app/engine/trade_outcome_refusal_log.go`)는 줄 쓰기를 **새 goroutine** 에서 한다(그 안에서도 recover).
  체결 트랜잭션은 SQLite 쓰기 잠금을 쥐고 있고, 그동안 손절 제출 기록도 그 잠금을 기다린다 — 막힌 로그 쓰기가 손절을 늦추면 안 된다
  (불변식 4). 비용은 거절 한 건당 goroutine 하나(포지션 종결마다 최대 한 번).
- 줄: `obs.Logger.Warn("trade_outcome.not_frozen", "symbol", …, "market", …, "instance_seq", …, "reason", …)`.
  `NewContext` 가 `opts.Logger`(생산 = engine.log 로거, `cmd/tossctl/engine_assembly.go`)로 바인딩한다. `bindApplyHooks(j)` 는
  시험 호출자 14곳을 위해 시그니처를 유지하고 `bindApplyHooksLogged(j, nil)` 에 위임한다.
- 줄은 트랜잭션 커밋 전에 예약된다. 커밋이 실패하면(뒤따르는 훅 없음 — Exit 이 마지막) 그 종결이 재적용될 때 줄이 한 번 더 날 수 있다.
  줄의 의미는 「이 종결 시도가 성과를 동결하지 못했다」 이다.

## D4. backfill 도구 — 미리보기 → 사람 확인 → 실행 (3항)

- 저널 쪽(새 파일 `trade_outcome_backfill.go`): `tradeOutcomeBackfillPlan(ctx, reader, model, now)` 이 CLOSED 포지션 전부를 돌며
  항목(종목 · 시장 · 회차 · 종결 시각 · 행 존재 여부 · 계산 결과 또는 거절 사유 · 기존 행이면 재계산 일치 여부)을 만든다.
  보존 기간(180일) 밖에서 종결된 포지션은 채움 대상에서 뺀다(보존 정리가 지운 행을 되살리지 않음).
  - `PreviewTradeOutcomeBackfill(ctx, *ReadOnly, model, now)` — 읽기 전용 연결(`mode=ro` + `query_only`)로 계획을 만든다.
    `*ReadOnly` 의 메서드 집합(쓰기 메서드 0 을 재는 readonly_test)은 늘리지 않는다 — 패키지 함수다.
  - `(*Journal).BackfillTradeOutcomes(ctx, model, now, token)` — 계획을 **다시** 계산해 토큰이 같을 때만 채움 대상마다 기존
    `BackfillTradeOutcome`(덮어쓰기 거절 · 동결 불변)을 부른다. 토큰이 다르면 아무것도 쓰지 않는다.
  - 토큰 = 채움 대상(포지션 id · 손익 · R · 수량 · 위험 · 비용)의 정규 직렬화 SHA-256 앞 16 hex. 미리보기와 실행 사이에 무엇이든
    바뀌면 토큰이 달라진다.
- 도구 `tools/a130-outcome-backfill`: 기본 = 미리보기(출력: 종목 · 시장 · 회차 · 종결 시각 · 판정 · 수량 · 실현 R · 손익 · 사유,
  기존 행 재계산 일치 수, 토큰). `--execute --confirm <토큰>` = 실행. 비용 모델은 엔진과 같은 `costs.DefaultModel()`.
  계좌 참조 · 포지션 id 는 출력하지 않는다. ⚠ mutating — 에이전트 · 시험은 운영 저널에 실행하지 않는다(시험은 임시 저널).

## D5. 콘솔 안내문 (4항)

`templates_portfolio.go` 의 「이 표의 한계」 안내문을 실제 누락 사유 넷(엔진이 관리하지 않던 포지션 · 외부 매도 · 대사 조정으로 닫힘 ·
엔진 매도 수량 ≠ 매수 쪽 수량 · 체결 · 조정 평균가 미관측)과 거절 줄 이름 · backfill 도구로 고친다. 기존 시험이 고정한 문장
「외부 매도로 종결된 포지션은 여기에 행이 생기지 않으므로」 는 그대로 둔다(그 문장은 여전히 참이다).

## D6. 잔여 (명명)

- 늘어난 수량의 매입가 = 조정 평균가(사용자 결정). 사람이 다른 값에 추가 매수하면 한계 매입가와 다르다(실측 11건엔 없음).
- 같은 초의 편입 뒤 조정은 빠져 거절된다(행 없음 + 사유 — 조작된 행은 아님).
- 엔진 줄은 goroutine 이라 프로세스가 곧바로 끝나면 줄이 유실될 수 있다(원장이 정본 — backfill 미리보기가 같은 사유를 다시 계산함).
- 기존 엔진 진입 경로의 `round_trip_unpriceable` 은 내부 갈래(진입 체결 없음 · 미관측 평균가 · 읽기 실패)를 하나로 묶는다.

## D7. 시험 (RED 먼저)

1. 편입 3 → 편입 뒤 조정 +3 → 엔진 6주 매도: 행이 생기고 initial_quantity = 6, R = pnl ÷ (r × 6) 정확값(2배 부풀림 아님 단언).
2. 편입 전 조정이 편입 수량에 이미 든 경우: 조정을 다시 더하지 않음(행 생김, 수량 = 편입 수량).
3. 편입 뒤 조정 평균가 미보고 → 행 없음, 사유 `adjustment_unpriced`.
4. 교차 경우(편입 뒤 +3 · 엔진 3주만 매도 · 사람이 나머지) → 행 없음, `sell_quantity_mismatch`.
5. 거절 싱크: 거절된 종결에서 정확히 1회, 사유 · 종목 · 시장 · 회차, 계좌 센티널 · 포지션 id 미포함. 성공 종결엔 0회.
6. 싱크 패닉 → 체결 · CLOSED · exit 완료는 커밋됨(비롤백). 싱크 nil → 거절 경로의 읽기 0(동작 이전과 같음).
7. 엔진 싱크: nil 로거 → nil, 줄의 키 집합 = {symbol, market, instance_seq, reason}(+ 공통 키), 쓰기는 호출 goroutine 밖(막힌 writer 가
   호출을 붙잡지 않음), `NewContext` 가 `bindApplyHooksLogged(jrn, opts.Logger)` 로 바인딩(구조 핀).
8. backfill: 미리보기는 쓰기 0(파일 바이트 · `total_changes` 불변), 토큰 불일치 실행은 0 행, 일치 실행은 대상 행만, 기존 행 재계산 일치.
9. 회귀: 기존 trade_outcomes · adoption_outcome 시험 전부 green.
