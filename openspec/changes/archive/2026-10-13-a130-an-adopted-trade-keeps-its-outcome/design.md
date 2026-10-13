# a130 design (4판)

base `233d5d9c`. 분기 주장은 전부 `analysis/function-logic/*/ast.json`(편집 전 추출) 의 열거에서 온다.

판 이력: 1판 REJECT(F-1~F-13) → 2판 「거절만」(R2-1~R2-11 APPROVE-WITH-FIXES) → 사용자 철회 → 3판 「조정 뒤 엔진 관측가」 REJECT(R3-1:
그 관측은 원장에 저장되지 않음 — 평평한 선은 이벤트 없이 덮어씀) → **4판: 사용자 결정 2026-10-13 3차 — 브로커 공식 1분봉(D 출처).**
사유만 기록(C) · 엔진 관측 저장(A) · 선 변경 이벤트(B)는 기각.

## 사실 (편집 전 AST · HEAD · 저널 사본 · 브로커 읽기 실측)

- `freezeTradeOutcomeTx`(trade_outcomes.go:129)의 유일한 호출자는 `ApplyExitFill` B14 안 :1023 이고 반환 bool 을 버린다.
- `computeTradeOutcome`(:163, 분기 12): 거절 return 9개. B5 :197 은 base 커버리지 0(외부 포지션은 `ApplyExitFill` B3 에서 먼저 돌아감).
- `adoptedRoundTripLegs`(:488): 매수 다리 = 편입 수량 × 편입 관측가, B3 엔진 매도 0, B5(:512) 매도 ≠ **편입 시점 수량** → 거절.
- 대사 수렴(`reconcile/converge.go:221`)은 평균가를 보내지 않고 `ApplyPositionAdjustment`(:312)는 이전 평균가를 이어 적는다 — 늘어난 수량의 매입가는 원장에 없다(F-1).
- 엔진의 5초 관측은 exit 선이 바뀔 때만 `exit_events` 에 남는다(평평한 선은 `RefreshExitObservation` 이 이벤트 없이 덮어씀, a111) — 원장 안에는 조정 순간의
  시장가가 없다(R3-1).
- 공식 `GET /api/v1/candles`: interval `1m`, `count` ≤ 200, `before` = 포함 상한, `adjusted` 플래그. 분봉 `timestamp` 는 **봉 종료 시각**이고 t 봉은 [t−1분, t)
  를 덮는다(internal/official/candle_reads.go 주석, a112 결정 30 · 31 실측). 2026-10-13 읽기 실측: 사건 조정 00:18:54Z 는 **00:19:00Z(09:19 KST) 봉**
  (open 18540 · close 18430)에 든다 — 09:18 KST 라는 봉은 09:17~09:18 을 덮으므로 그 분이 아니다.
- 기존 읽기 메서드: `Client.Candles`(float64 로 바꿈) · `RawMinuteCandles`(KR 전용) · **`StrictMinuteCandles`**(strict_minute_candles.go:131 — KR/US, 소수 문자열 그대로,
  `adjusted=false`, 응답 문법 · 통화 · 소수 · 분 경계 검증, before 는 숫자 오프셋). 4판의 1분봉 읽기는 이것을 count=1 로 감싼다(리뷰 4차 P2-5 — 1판의 「없다」 는 틀렸음).
- `trade_outcomes` 에는 가격 출처 칼럼이 없다.

## D1. 편입 포지션의 매수 다리 — 묶음별 브로커 1분봉 (사용자 결정 2026-10-13 3차)

```text
q_a   = position_adoptions.quantity           편입 수량
p_a   = position_adoptions.observed_price     편입 관측가
t_a   = position_adoptions.observed_at        편입 관측 시각
r     = exit_states.initial_risk              주당 최초 위험 — 편입 시점 값, 불변(= p_a − synthetic_stop)

lot k = position_adjustments 행 중 position_id = 이 포지션, created_at_k > t_a, new_k > prev_k
Δ⁺_k  = new_k − prev_k
L_k   = floor_minute(created_at_k) + 1분      그 조정 시각이 든 1분봉의 이름(봉 종료 시각)
p_k   = 브로커 공식 1분봉 L_k(adjusted=false)의 종가. 그 이름의 봉이 없으면 거절(no_intake_candle)

Q_buy = q_a + Σ_k Δ⁺_k                         이 왕복의 초기 수량
N_buy = q_a·p_a + Σ_k Δ⁺_k·p_k
동결 조건: 재편입 없음 ∧ 편입 뒤 감소 없음 ∧ 모든 lot 에 p_k ∧ 인스턴스 매도 무결(D1.3) ∧ sell.priced ∧ Q_sell = Q_buy
pnl   = N_sell − N_buy − cost_buy(N_buy) − cost_sell(N_sell)
R     = pnl ÷ (r × Q_sell)          분자 · 분모가 같은 Q_sell(= Q_buy) — 분모를 q_a 로 두면 R 이 부풂
```

1. **A7 과의 정합**: 편입 규칙(기준선은 매입 원가가 아니라 편입 시점 관측가)을 묶음마다 같은 규칙으로 적용한다. 출처가 엔진 관측(Quote.Last)에서 브로커 시장
   분봉으로 바뀔 뿐 「그 시점 시장가」 라는 뜻은 같다. cost_basis · 조정 행 평균가는 쓰지 않는다.
2. **봉 안의 어느 값인가 — 종가**: A7 의 p_a 는 관측 순간의 마지막 체결가(Quote.Last)다. 봉의 종가는 그 분의 마지막 체결가로 같은 종류의 값이다. 시가는
   그 분의 첫 체결, 고가 · 저가는 극값이라 「그 순간의 가격」 이 아니다. 조정 시각과 종가의 시차는 1분 이하다(분봉 해상도의 한계 — D5).
3. **`adjusted=false`**: 매도 다리는 그날 실제 체결 평균가(조정 안 된 값)다. 수정주가는 그 뒤 액면분할 · 배당으로 과거 값을 다시 쓰므로 같은 날의 실제 체결가와
   비교할 수 없다. 원장의 손익은 실제 거래 가격끼리여야 한다.
4. **정규장 밖 조정(실측)**: 모집단 12건의 편입 뒤 증가 묶음 23개 — 정규장 안 17 · 밖 6(KR 장 전 08:29 KST 1 · KR 장 후 15:40 · 15:46 · 20:02 KST 3 ·
   US 장 후 17:09 · 19:03 ET 2). 같은 분 봉 실측: 23 중 **21 있음**(정규장 17 + 정규장 밖 4 — NXT 장 전 · 장 후, US 장 후 체결이 있는 분), **2 없음**(15:40 KST
   — KRX 마감 뒤 NXT 체결 없음, 20:02 KST — NXT 마감 뒤). 결정: **그 분의 봉만 받는다.** 직전 정규장 마지막 봉을 허용하면 15:40 은 10분 전, 20:02 는 2분 전
   가격을 쓰게 되고, 그 시각 수량이 언제 산 것인지는 원장에 없다 — 가격을 지어내는 쪽이므로 받지 않는다(`no_intake_candle`).
5. **인스턴스 매도 무결(R3-3 · 리뷰 4차 P2-3 · P3-8)**: 귀속된 매도 주문 중 이 인스턴스가 열리기(opened_at) 전 **또는 닫힌(closed_at) 뒤**에 확정된 것이 있으면
   `foreign_instance_sell`(심볼 단위 flatten 갈래가 이전 · 이후 인스턴스 매도를 쓸어 담음 — 이후 쪽은 backfill 이 늦게 돌 때), **편입 관측 시각 뒤**(엄격) ·
   닫히기 전에 확정된, 체결 있는 같은 계좌 · 시장 · 심볼의 로컬 매도 주문이 귀속 밖이면 `unattributed_local_sell`(사람이 낸 로컬 매도; fold-in 과 편입 사이의
   매도는 편입 수량에 이미 반영돼 세지 않음). 시각은 주문 확정 시각(`mutation_attempts.settled_at`, 원장 시계)이고 체결 `committed_at` 이 아니다 —
   committed_at 은 소유 확정보다 엄격히 뒤로 밀려 적힐 수 있다(`journalTimeStrictlyAfter`). **시각은 거절 조건에만 쓰고 숫자를 귀속하지 않는다**(시간창
   귀속 금지 계약). 실측: 기존 고정 시계 시험(`TestOutcomesAreSplitByWhatJustifiedThePosition`)에서 committed_at 기준은 앞 인스턴스 매도를 1초 뒤로 읽어 헛거절,
   settled_at 기준은 통과.
6. **편입 뒤 감소(F-3)** `created_at ≥ t_a` → `post_adoption_disposal`. **이름 붙인 회귀(F-9)**: 편입 3 → +3 → −3 → 엔진 3주 매도는 a130 이전엔 동결, 뒤엔 거절.
   **재편입(I-2)** → `readopted` — 판정은 exit 상태 한 줄의 `lifecycle_generation > 1`(재편입이 READOPTED 이벤트와 같은 트랜잭션에서 씀). exit_events 를 훑는
   판정은 포지션 하나에 이벤트가 27,071 줄(사본 최대)이라 체결 트랜잭션 안 판정을 최대 20 ms 대로 늘렸음(리뷰 4차 P2-4 실측) — 한 줄 읽기로 바꿔 최대 2.4~2.9 ms
   (a130 이전 판정 최대 1.7~2.4 ms, 같은 사본).
7. **같은 초 경계**: 증가는 엄격(`>`, fold-in 0→q 와 편입 직전 증가는 편입 수량에 들어 있음), 감소는 포함(`>=`). 시각은 RFC3339Nano 로 파싱해 비교한다(문자열 비교 아님).
8. 편입 뒤 증가가 없으면 Q_buy = q_a, N_buy = q_a·p_a — **a130 이전 산식과 같다**(사본 동결 31행 재계산 31/31).

## D2. 사유 코드 — 판정 함수가 돌려주는 타입

`computeTradeOutcome(ctx, r, positionID, model, prices IntakePrices)` · `adoptedRoundTripLegs(…, prices)` 는 `TradeOutcomeRefusalReason`(새 파일
`trade_outcome_refusal.go`, 빈 값 = 성공)을 돌려준다. 상수 20개가 전부다(시험이 go/parser 로 선언과 목록 대조).

| 코드 | 자리 | 체결 트랜잭션 줄에 나올 수 있나 |
|---|---|---|
| `cost_model_unconfigured` · `position_unreadable` · `empty_buy_leg` · `cost_estimate_failed` | computeTradeOutcome | 예 |
| `unmanaged_position` | :197 | 아니오(exit 상태 없음 → ApplyExitFill B3) — census · backfill 만 |
| `exit_state_unreadable` | :202 | 예 — `PendingState` 는 읽혔지만 `frozenExitFacts` 의 읽기 · 해석이 실패한 경우(리뷰 4차 P3-9) |
| `round_trip_unpriceable` | 결정 경로 | 예(진입 체결 없음 · 미관측 평균가 · 읽기 실패를 묶음) |
| `adoption_unreadable` · `readopted` · `post_adoption_disposal` · `adoption_epoch_unreadable` | 편입 기록 · 편입 시기 | 예 |
| `needs_intake_price` | 늘어난 묶음이 있는데 분봉 집합이 nil(체결 트랜잭션 — 네트워크 금지) | **예 — 4판의 주 거절** |
| `no_intake_candle` | 그 분의 봉이 없음 | 아니오(backfill 만) |
| `engine_sell_unreadable` · `no_engine_sell` · `sell_unpriced` · `foreign_instance_sell` · `unattributed_local_sell` · `sell_quantity_mismatch` | 매도 다리 | 예 |
| `insert_failed` | freezeTradeOutcomeTx | 예 |

## D3. 체결 트랜잭션 — 네트워크 0, 거절은 커밋 뒤에만, 계좌 없이

- 체결 트랜잭션 안의 판정은 `prices = nil` 로 돈다 — 늘어난 묶음이 있는 편입 포지션은 `needs_intake_price` 로 거절되고, 가격 산정과 행 쓰기는 backfill 몫이다
  (사용자 결정 3: 분석이 체결 반영을 막거나 늦추면 안 됨).
- `ApplyHooks.OutcomeRefused func(TradeOutcomeRefusal)`. `TradeOutcomeRefusal` = {Symbol, Market, InstanceSeq, Reason} — 계좌 참조 · 포지션 id 없음.
- `ApplyExitFill` B14: 사유가 있으면 `tx.noteOutcomeRefusal` — 싱크가 묶였고 ctx 에 수집 목록이 있을 때만 `positions` 한 줄을 읽어 쌓는다. 오류를 돌려주지 않는다.
- `RecordFill` 이 B37 에서 `withOutcomeRefusals(ctx, &refusals)` 를 넘기고 `tx.Commit()` 성공 뒤에만 `j.reportOutcomeRefusals` — 싱크 호출마다 recover. `runApplyHooks`
  시그니처 그대로(핸들 리터럴 필드 하나), 전략 되메움 경로 무편집(매수만).
- 엔진 싱크는 줄 쓰기를 새 goroutine 으로 넘긴다(체결 감지 루프 비차단). 줄은 **best-effort** — 프로세스가 곧바로 끝나면 유실될 수 있다(원장이 정본; R3-10).
  `NewContext` 가 `bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger))`, `bindApplyHooks` 는 선택 인자(시험 14곳 그대로).
- 편입 시기 읽기 · backfill 모집단은 `positions.adoption_id` 를 쓴다 — `adoption_static_test.go` 읽기 허용 목록에 두 파일(UPDATE 없음).

## D4. backfill — 분봉은 도구가 읽고, 원장은 미리보기 → 사람 확인 → 항목별 재대조

- **모집단**: `state = 'CLOSED'` ∧ 편입 기록 ∧ `COMPLETED` exit 이벤트(체결로 닫힘 — 대사로 닫히면 `ADJUSTMENT_CLOSED` 가 completed=1 을 먼저 써서 COMPLETED 는
  생기지 않음, 둘은 배타적) ∧ 성과 행 없음. 보존 기간(180일) 밖은 `outside_retention` 표시만. 엔진 진입 포지션 제외(F-2). 엔진 귀속은 판정의 매도 검사가 강제.
- **⓪ 묶음 목록** `IntakeLotsForBackfill(ctx, *ReadOnly)` — 늘어난 묶음마다 시장 · 종목 · 봉 이름(L_k) · 불투명 열쇠(포지션 id 를 감춘 해시).
- **분봉 읽기(도구)**: `official.Client.MinuteCandleClosingAt(ctx, market, symbol, label)` — `StrictMinuteCandles(market, symbol, 1, before = label 의 서울 오프셋 표기)`
  를 감싸고, 돌려받은 봉의 이름이 label 과 **같을 때만** 종가(소수 문자열 그대로)와 **돌려받은 이름**을 돌려준다. 도구는 그 이름을 `IntakeCandle.Label` 로 적고
  원장 판정 `priceLots` 가 묶음의 분과 다시 대조한다(두 번째 가드 — 리뷰 4차 P1-2). 생산 client 는 a087 · a128 과 같은 방식으로 열되 **`--config-dir` 를 엔진의 config
  디렉터리로 꼭 넘겨야** 한다: 엔진은 `--config-dir`(compose 의 TOSSOS_CONFIG_DIR)로 돌고 토큰 캐시가 그 안에 있다. 빼면 apppaths 기본값(사용자 캐시 디렉터리)의
  다른 토큰 파일로 새 토큰을 발급받아 엔진 토큰을 무효화한다(a082 — 리뷰 4차 P1-1; 2026-10-13 실측: 이 change 의 기본값 미리보기 직후 두 토큰 파일이 2초 차이로 다시
  쓰였음). 도구는 묶음이 있는데 `--config-dir` 가 없으면 client 를 열기 전에 거절하고, 쓰는 토큰 파일 경로를 출력한다.
  **예산**: candles 는 초당 5회 묶음(internal/candidate veto.go 주석); 도구는 1.2초에 1회, 429 · 오류는 재시도 없이 쓰기 전에 멈춘다. 모집단 12 → 23회(약 28초).
- **① 미리보기** `PreviewTradeOutcomeBackfill(ctx, *ReadOnly, model, now, prices)` — 쓰기 0. 항목마다 판정 · 계산 결과 · 쓴 분봉(이름 · 종가) · 거절 사유, 확인 토큰
  (채움 대상 값 + 쓴 분봉 + 비용 모델(`%#v`) + 빌드 스키마 버전의 SHA-256 앞 16 hex, R3-13).
- **② 실행** `(*Journal).BackfillTradeOutcomes(ctx, model, now, prices, token)` — 분봉을 다시 읽어(트랜잭션 밖) 계획을 다시 계산, 토큰이 다르면 0 행(분봉 값이
  바뀌어도 다름). 같으면 채움 대상마다 짧은 BEGIN IMMEDIATE 하나: 행 존재 확인 → 같은 분봉으로 재계산 → 계획 값과 모든 칸 대조(다르면 `ErrBackfillItemMoved`,
  그 항목과 뒤는 안 씀) → INSERT → 커밋. 트랜잭션 안에 네트워크 없음.
- **쓰기 잠금 시간(R3-5, 실측)**: 이관한 사본(스키마 35)에서 9 항목 실행 — 항목당 9.8~22.2 ms(최대 22.17 ms). 손절 기록이 기다릴 수 있는 상한이 항목당 이 정도다.
  그래도 실행은 두 시장이 닫힌 시간에 한다(KST 기준 창은 미국 서머타임에 따라 달라짐 — 운영 절차가 그날의 창을 정함, 리뷰 4차 P3-11) — 「분석 경로의 격리」.
- **스키마 가드 순서(R3-4)**: 읽기 전용으로 열기 → `SchemaVersion()` = 빌드(35) 확인 → 같을 때만 `journal.Open`(이관 · 무결성 검사를 함). 다르면 쓰기 연결을 열지
  않는다(시험: 스키마 32 파일 실행 거절 뒤 바이트 동일).
- **출처 기록**: `trade_outcomes` 에 출처 칼럼이 없고 **스키마를 바꾸지 않는다**. 근거: ① 출처는 유도 가능하다 — 편입 뒤 증가 묶음이 있는 편입 포지션의 행은
  이 도구로만 생긴다(체결 트랜잭션은 그 경우 늘 `needs_intake_price`). ② 쓴 분봉은 다시 읽을 수 있다(과거 1분봉, adjusted=false). ③ 스키마 변경은 이미지
  롤백을 백업 복원으로 만든다(운영은 10-14 에 32→35 이관 예정 — 그 위에 36 을 더 쌓지 않음). 도구 출력이 항목마다 분봉 이름 · 종가 · 출처(`broker_candle_1m`)를 보인다.
- **사본 미리보기(실측, 저장소에는 수 · 사유만)**: 모집단 12 → fill 9 · `no_intake_candle` 2 · `sell_unpriced` 1. 묶음 23 · 분봉 있음 21 · 없음 2.
- **집계의 세 번째 출처(리뷰 4차 P3-13)**: backfill 이 쓴 행(편입 뒤 증가 묶음이 분봉으로 매겨진 편입 행)은 유도 가능하다 — 편입 포지션이면서 `created_at > observed_at` ·
  `new > prev` 인 조정이 있는 행. 「합성 R 의 구분 집계」 는 그것을 편입(합성 R) 쪽으로 세며, 따로 나눠 보려면 그 조건으로 가른다.
- 도구는 `--journal` 경로 필수(기본값 없음), 비용 모델은 엔진과 같은 `costs.DefaultModel()`, 출력에 계좌 참조 · 포지션 id 없음. ⚠ mutating — 사람만 실행.

## D5. 잔여 (명명)

- 묶음 편입가는 그 분의 마지막 체결가다 — 그 묶음의 실제 매입가가 아니다(실제 매입가는 a132 이후 converge 가 기록하기 전엔 원장에 없다). 조정 시각과 종가의
  시차는 1분 이하.
- 앞으로 생기는 편입 뒤 증가 종결은 체결 트랜잭션에서 `needs_intake_price` 로 거절되고 사람이 backfill 을 돌릴 때까지 행이 없다 — 주기적 읽기 전용 자동화는
  범위 밖(issues.md I-4 후속).
- 인스턴스 매도 무결의 같은 초 경계: 편입 관측과 같은 초에 확정된 귀속 밖 매도는 못 보고 매도 수량 검사만 남는다. 엔진 BUY 추가 매수 · 모호한 소유는 수량 불일치 ·
  귀속 밖 매도로 거절된다(fail-closed). backfill 은 늦을수록 뒤 인스턴스 flatten 과 겹쳐 `foreign_instance_sell` 로 거절될 수 있으니 배포 뒤 곧 돌린다. `AdoptPosition` 은 요청 수량이 그 순간
  인스턴스 수량과 같은지 확인하지 않는다(감소 규칙의 시각 경계는 fold-in · 수렴 · 편입이 한 goroutine 에서 차례로 도는 것에 기댐).
- 일시 실패와 구조 거절이 섞인 사유(`round_trip_unpriceable` · `*_unreadable`). `roundTripLegs`(결정 경로)의 트랜잭션 밖 부정확(F-2)과 생산 호출자 0 인
  `BackfillTradeOutcome` 의 그 위험은 그대로 — a130 backfill 은 그 경로를 쓰지 않는다.
- 엔진 줄은 best-effort. 목록 없이 싱크만 있는 `runApplyHooks` 호출자는 거절을 버린다.
- **인스턴스 매도 무결 읽기의 규모(구현 리뷰 P3-1)**: 로컬 매도 질의는 `UPPER(TRIM(f.symbol))` · `TRIM(f.account_ref)` 로 비교해 체결 표 색인을 못 쓴다(전체
  훑기, 기존 `engineSellOrders` · `sumSellFills` 와 같은 꼴). 리뷰 실측: 합성 체결 20만 행에서 종결 하나의 판정 102 → 198 ms. 운영 체결 행은 67(사본 실측,
  2026-10-13) — 판정 최대 2.694 ms. 체결 표가 수만 행에 이르면 체결 트랜잭션의 쓰기 잠금 시간에 더해지므로(불변식 4 쪽 비용) 그때 정규화 값 등식 비교로
  바꾼다 — 지금은 바꾸지 않는다(정규화 저장 여부를 전수 확인해야 하는 별도 변경).
- `BackfillTradeOutcome`(기존 단건 수동 채움, 생산 호출자 0)은 모집단 한정이 없고 오류 문구에 포지션 id 를 싣는다(a130 이전부터). a130 은 거절 사유만 문구에
  더했다. 제거 · 한정은 별도 변경(시험 고정 다수).

## D6. 콘솔 안내문

`templates_portfolio.go` 「이 표의 한계」: 성과 행은 엔진이 관리하던 포지션이 엔진 매도로 닫히고 수량 · 가격이 맞을 때만 동결된다는 것, 행이 없는 실제 사유
(엔진이 관리하지 않던 포지션 · 외부 매도 · 대사 조정으로 닫힘 · 편입 뒤 늘어난 수량은 그 시각 브로커 1분봉으로 다시 계산해야 함(backfill 도구) · 편입 뒤
수량이 줄었거나 재편입됨 · 매도 수량 불일치 · 평균가 미관측), 거절 사유가 `engine.log` 의 `trade_outcome.not_frozen` 줄에 남는다는 것. 고정 문장 유지.
바꾼 화면은 실제로 렌더해 확인한다(규칙 13).

## D7. 시험 (RED 먼저) · 변이

시험: `internal/journal/a130_adopted_outcome_test.go` · `a130_trade_outcome_backfill_test.go` · `a130_journal_copy_census_test.go`(env) ·
`internal/app/engine/a130_trade_outcome_refusal_log_test.go` · `internal/console/a130_history_note_test.go` · `tools/a130-outcome-backfill/main_test.go`.

1. 체결 트랜잭션: `TestA130AGrownPositionIsRefusedInsideTheFillTransaction`(needs_intake_price · 커밋 뒤 1회 · 계좌 없음).
2. 산식: `TestA130AGrownPositionPricesEachLotAtItsMinuteBar`(3·70000 + 3·69000, R = pnl ÷ (3500 × 6), 2배 아님, 조정 평균가 아님) ·
   `TestA130AMissingOrWrongMinuteBarIsRefused`(없음 · 앞 분 봉) · `TestA130TheBarLabelIsTheMinuteThatClosesAfterTheAdjustment`(00:18:54→00:19:00 · 정각→다음 분 · 소수 초 · 시간대).
3. 거절: F-9 · F-3 · 재편입 · `TestA130AnUnattributedLocalSellIsRefused` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` ·
   `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell` · `TestA130AnUnpricedEngineSellIsReportedAsSuch`.
4. 불변: 편입 전 조정 · fold-in(30초 앞 · 같은 초) · 같은 초 경계 · 엔진 진입 · 기존 adoption · trade_outcomes · trade_analytics 시험.
5. 기록: 롤백 무보고 · 싱크 패닉 비롤백 · RecordFill 구조 핀(보고는 마지막 Commit 뒤) · 비용 모델 미구성 사유 · 자리별 사유 대조 · 사유 집합 20 · 필드 넷.
6. backfill: 모집단(편입 · 체결 종결 · 행 없음만), 묶음 목록 순서, 미리보기 쓰기 0(db · -wal), 토큰 불일치 · 분봉 값 변경 → 0 행, 일치 → 미리 본 행, 재사용 0,
   항목 이동 → 안 씀, 보존 밖.
7. 도구: 경로 · 확인 없이 거절, 묶음 없으면 브로커 client 를 열지 않음, 스키마 32 실행 거절 + 바이트 동일, 분봉은 그 분만 · 간격 두고 읽음.
8. 구현 리뷰 보강(2026-10-13): 종가 0 봉 거절(`TestA130AZeroCloseBarIsRefused`, 변이 M32) · 싱크 없음 가드 AST 핀(`TestA130ANilSinkReadsNothingOnTheRefusalPath`,
   M34) · 실패한 정정 시도가 로컬 매도 시각을 앞당기지 않음(`TestA130AFailedAmendDoesNotDateALocalSell`, M33) · 분봉 행의 유도 구분
   (`TestA130BarPricedRowsAreDerivableFromTheLedger`). 원장 의존 그래프 검사는 `net/http` · `internal/official` 만 거름 — `net` 은 기존 의존
   (github.com/google/uuid · crypto/x509)이 이미 끌어와 거르면 a130 이전 트리도 실패(측정).
8. 엔진 · 콘솔: 싱크 nil · 줄 키 · 막힌 writer · NewContext 구조 핀 · 안내문 문구.

변이(사본 · 무변이 대조군 GREEN 먼저, `analysis/harness/a130_mutate.py`): 묶음 무시 · 봉 이름 앞 분 · 다른 분 봉 허용 · nil 분봉에서 묶음 무시 · 증가 엄격 `>`→`>=` ·
감소 검사 삭제 · 감소 `>=`→`>` · 재편입 삭제 · 매도 비교를 편입 수량으로 원복 · 묶음 대금 누락 · 분모 바꿔치기 · foreign 검사 삭제 · unattributed 검사 삭제 ·
settled_at → committed_at · 거절 기록 삭제 · 커밋 전 보고 · 싱크 recover 삭제 · 엔진 goroutine 제거 · NewContext 싱크 인자 삭제 · 핸들 싱크 필드 삭제 ·
모집단 편입 필터 삭제 · 항목 재대조 삭제 · 토큰에서 분봉 제외 · 스키마 가드 제거 · sell_unpriced → mismatch.
