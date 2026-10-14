# a137 design — 일일 손실은 측정된 값이다

> 상태: **사람 결정 대기.** 아래 D1~D6 은 권장안이고, 「사람 결정」의 Q1~Q8 이 정해지기 전에는 base 를 고정하지 않고
> production 파일을 편집하지 않는다(팀메이트 지시 「사람 결정이 필요하면 proposal/design 뒤 멈춤」).

## D0. 현재 동작 — 증거

분기 주장은 전부 `analysis/ast-evidence/`(tools/logic-map, HEAD `7b073f55`)의 열거에서 왔다. 호출 관계는 codegraph 1.6.0
(`analysis/code-context/raw/`). 요약은 proposal.md 「지금 Guardian 이 일일 손실로 하는 일」 1~6 이고, 여기서는 설계가 의존하는 사실만 다시 적는다.

| # | 사실 | 근거 |
|---|---|---|
| F1 | `checkDailyLoss` 는 입력이 없거나 음수면 `INPUT_UNAVAILABLE`, 한도 이상이면 `DAILY_LOSS_LIMIT_REACHED`, 0 이면 통과 | `risk--checkdailyloss.ast.json` B4~B9 · `risk--magnitudein.ast.json` B1~B3 |
| F2 | 매도 의도는 `entryChain` 에 들어가지 않는다 | `risk--evaluate.ast.json` B2(135) → `evaluateReduction` |
| F3 | `riskcalc.DailyLoss` 생산 호출자 0 | `raw/callers-DailyLoss.txt`(6, 전부 시험) |
| F4 | `riskcalc.DailyLoss` 는 모든 항목의 `TradingDay` 가 측정일과 같고 · `CostsDeducted` 이고 · `AsOf` 가 계좌 스냅샷 신선 한계(10초) 안이고 · 외화면 신선한 환율(60초)이 있어야 값을 낸다. 아니면 unknown 오류 | `riskcalc--dailyloss.ast.json` B1~B9 |
| F5 | 체인에 값을 넣는 생산 자리: tracer 리터럴 `"0"`(생산 생성자 0) · 서명 계좌 매니페스트 `daily_realized_loss_base`(저장소 안 생산자 0) | `engine--tracer.accountstate.ast.json` · `strategyaccount--sealauthority.ast.json` B1 · grep |
| F6 | 자동 강화(ENTRY_BLOCKED)는 `IssueEntry` 에만 있다. 생산 전략 경로(`PrecheckQFinalEntry` B16)의 한도 거절은 강화 없이 `AUTHORITY_MISMATCH` 결함이 된다 | `raw/callers-escalateFor.txt` · `execgw--riskguardian.precheckqfinalentry.ast.json` · strategy_first_leg_admission.go:93 |
| F7 | 원장의 거래일 경계 관례: 일일 손실 예약은 **결정의 시장** 현지 날짜(`clock.Market.TradingDay`)로 소멸한다 | reservation_release.go:379-455 `sweepTradingDayBoundary` · clock/market.go:158 |
| F8 | `trade_outcomes` 행은 포지션이 **완전히 닫힐 때** 체결 트랜잭션 안에서 한 번 동결된다. 동결 거절 사유 20 종(a130)은 원장에 남지 않고 싱크(로그)로만 간다 | trade_outcomes.go 머리 주석 · trade_outcome_refusal.go 상수 집합 |
| F9 | 공식 Open API 에 실현 손익 조회가 없다(endpoint 열거: accounts · holdings · orders · trades · prices · exchange-rate 등). WTS `client.GetDailyProfit` 은 있으나 엔진은 official-only 다 | `internal/official/*.go` 경로 문자열 · internal/client/profit_period.go:117 |

F6 의 하류(전략 감독자의 `latchMarket` 이 그 결함으로 시장 레인을 잠그는지, 그 잠금이 a112 5.3.3 의 durable latch 인지)는
HEAD 읽기로 `strategy_entry_supervisor.go:966` 까지 추적했고 **AST 산출물로 고정하지 않았다** — 구현 전 FLM 대상(tasks 0.7).

## D1. 측정기의 자리 — Guardian 이 바닥을 측정한다 (권장, Q6 과 묶임)

호출자 둘(F5)을 각각 고치는 대신 **Guardian 안에서** 진입 판정 직전에 원장 측정값을 계산하고,
`Account.DailyRealizedLoss = max(호출자 값, 측정값)` 으로 바꿔 체인에 넣는다.

- 한 자리(`RiskGuardian` 의 진입 입구 둘: `IssueEntry` · `PrecheckQFinalEntry`)가 두 생산 경로를 다 덮는다.
- 서명 매니페스트의 값은 **바닥으로만** 쓰인다 — 측정값이 크면 측정값, 서명값이 크면 서명값. 낮추는 경로가 없다.
- 측정 실패는 체인 앞에서 거절하지 않고 `Account.DailyRealizedLoss` 를 쓸 수 없는 값(통화 빈 Money)으로 두어 **`daily_loss` 단의 F1 B4**
  가 `INPUT_UNAVAILABLE` 로 거절하게 한다. 그래야 kill switch · 운영 모드처럼 앞 단의 정확한 사유가 가려지지 않는다
  (chain.go `Evaluate` 머리 주석 「첫 실패에서 정지 … 뒤 단의 입력은 쓸 수 없어도 된다」). 상세 문구에는 측정 실패 원인을 싣는다
  (계좌 값 · 원문 오류는 a131/a133 경계를 따름).
- `IssueReduction` 은 측정기를 부르지 않는다(F2 와 별개로 호출 자체가 없음을 시험으로 고정).
- 측정은 결정 · 예약 트랜잭션 **밖**에서 한다(order-execution 「원자적 위험 예약」 — 트랜잭션 안 브로커 조회 금지. Q5 가 환율 읽기를
  더하면 그것도 밖).

대안: (b) 각 호출자가 측정 — tracer 는 생산 생성자 0, 전략 경로는 서명 권한의 값을 덮어써야 해서 권한의 `identity` 의미가 바뀐다.
(c) 저장소 밖 매니페스트 서명자가 측정 — 저장소가 검증할 수 없다. 둘 다 권장하지 않음.

## D2. 측정 원천 — 원장 `trade_outcomes` (권장, Q1)

`TradeOutcome.RealizedPnLAfterCosts`(부호 있음, 두 다리 비용 차감)를 `positions` 와 조인해 계좌 · 시장 · `closed_at` 을 얻는다.
`riskcalc.RealizedPnL{TradingDay, Amount, Currency, CostsDeducted: true, AsOf}` 로 옮겨 `riskcalc.DailyLoss` 를 부른다 — 공식을 새로 만들지 않는다.

**계약 충돌 하나**: `riskcalc.RealizedPnL.AsOf` 주석은 「이 숫자는 브로커에서 오고 계좌 스냅샷 신선 한계를 따른다」이다(riskcalc.go:238-240).
원장 값에는 브로커 AsOf 가 없다. 권장: `AsOf = 원장을 읽은 순간` 으로 두고 riskcalc 주석을 「원장 원천일 때는 읽은 순간」으로 정정한다.
원장의 신선도(체결이 원장에 들어오는 지연)는 이 측정기가 아니라 기존 강화 트리거 「체결 감지 사이클 연속 5주기 실패 → ENTRY_BLOCKED」
(risk-management 운영 모드 요구)가 맡는다. 이 정정은 계약 문장 변경이므로 사람 승인 대상이다.

## D3. 원장이 세지 못하는 손실 — 보수/허용의 선택 (Q2 · Q3)

원장 원천은 엔진이 귀속한 왕복만 센다. 다음은 **세지 못한다**(허용 쪽으로 틀리는 것):

| # | 빠지는 손실 | 감지 가능? | 권장 |
|---|---|---|---|
| U1 | 외부 수동 거래(앱에서 사람이 판 것) | 원장에 없음 | a137 범위 밖 — order-execution 「외부 수동 거래는 브로커 스냅샷으로만」 · F9 로 공식 원천 없음. 명시적 한계로 적음 |
| U2 | 아직 열린 포지션의 부분 청산(사다리 rung · 부분 익절)으로 확정된 손실 | 오늘의 SELL 체결로 감지 가능 · 금액은 새 공식 필요 | **Q2** |
| U3 | 오늘 CLOSED 된 관리 포지션인데 성과 행이 없음(동결 거절 20 종) | `positions.closed_at` + 성과 행 부재로 감지 가능 | **Q3** — 권장: unknown → 진입 fail-closed |
| U4 | 브로커는 체결했으나 체결 감지가 아직 원장에 넣지 않은 것 | 불가(체결 감지가 원천) | 기존 5주기 트리거가 맡음 |
| U5 | 열린 포지션의 미실현 손실 | 가능 | **Q8** — riskcalc 계약이 제외(aggregate.go:206-209), 후속 change |

## D4. 일 경계 — 진입 시장의 현지 거래일 (권장, Q4)

기존 관례(F7)를 재사용한다: 측정일 = `intent.Market.TradingDay(now)`. 계좌의 **모든 시장**의 성과 중 `closed_at` 이
**그 시장 현지 날짜**에 드는 것을 센다(리스크 권위가 계좌 하나 — risk-management 「KR과 US는 하나의 account-base Guardian 권위」).
각 성과의 `RealizedPnL.TradingDay` 는 그 성과 시장이 아니라 **측정 시장**의 날짜로 붙인다 — riskcalc F4 의 「모두 같은 날」 검사를
만족시키는 유일한 일관 해석이다.

귀결(사람이 알아야 할 것): KR 진입과 US 진입은 서로 다른 24시간 창을 본다. 화요일 05:00 KST 에 닫힌 US 손실은 KR 화요일 창과
US 월요일 창 양쪽에 들어간다(이중 계상 — 보수 방향, 누락 없음). 콘솔 「오늘 실현손익」은 시장별로 나누고 합산하지 않는다
(operator-console) — 화면과 Guardian 의 숫자가 다를 수 있으므로 화면 문구 정정이 후속으로 필요할 수 있다.

## D5. 통화 — 외화 성과의 환산 (Q5)

한도는 계좌 기준 통화 하나(기본 KRW)다. KR 진입의 측정에 오늘 닫힌 US 성과(USD)가 들어가면 USD→KRW 환율이 필요하고,
riskcalc 는 신선한(60초) 환율이 없으면 unknown(F4) 이다. KR 진입 요청은 identity FX 만 가진다.

- (a) 같은 쌍의 봉인된 FX 권한이 요청에 있을 때만 환산, 없으면 fail-closed → **US 포지션이 닫힌 날 KR 진입 전부 거절**(정상 입력 거절).
- (b) 기존 공식 환율 읽기(`officialfx` 경로 — 새 전송 아님)를 측정마다 한 번 → 진입당 브로커 호출 1 추가, retry matrix · rate budget 계상 필요(§0.4).
- (c) 측정 시점 직전의 가장 최근 봉인 환율에 보수 haircut — 새 숫자 · 출처 등재 필요.

권장: (b). 단 「새 브로커 전송 0」 지시와의 경계(기존 읽기 재사용 = 새 전송 아님)와 호출 예산은 사람이 확인.

## D6. 전략 경로의 분류와 강화 (Q6 · Q7)

측정이 실제 손실을 넣기 시작하면 F6 이 바로 살아난다: 생산 전략 경로에서 한도 도달은 결함(`AUTHORITY_MISMATCH`)이 되어 주기를 멈추고
감독자가 시장 레인을 잠글 수 있다(D0 미고정 하류). 하루 한도 도달은 결함이 아니라 정책 결과다.

권장: `PrecheckQFinalEntry` 의 체인 거절 중 `DAILY_LOSS_LIMIT_REACHED` 는 `IssueEntry` 와 같은 `escalateFor` 를 거쳐 ENTRY_BLOCKED 로
강화하고, 호출자는 이것을 레인 결함이 아닌 계좌 단위 진입 차단으로 분류한다. 레인 잠금 의미가 바뀌므로 사람 결정.

Q7: 자동 강화는 durable 이고 완화는 사람 승인뿐이다(risk-management). 그래서 한도 도달은 **그날로 끝나지 않고** 사람이 풀 때까지
진입을 막는다 — 일일 손실 예약이 거래일 경계에서 소멸하는 것과 다르다. 스펙대로 유지할지 확인.

## Fail-closed 가 거부하는 정상 입력 (권장안 기준, 기억 규칙 「fail-closed는 무엇을 거부하는지 말해야 한다」)

진입만 거부한다. 매도 · 손절 · 비상 청산 · 대사 · 체결 감지는 측정기를 부르지 않으므로 어느 것도 거부되지 않는다.

1. 원장 읽기 오류(SQLite busy · 잠금 · 닫힌 핸들) — 그 시도 한 번의 진입. 일시적.
2. 오늘 닫힌 관리 포지션 중 성과 행이 없는 것(U3) — **그 시장일의 나머지 진입 전부**, 사람이 `tools/a130-outcome-backfill` 로 되메울 때까지.
   예: 편입 포지션을 엔진이 팔았는데 편입가 분봉이 없음(`needs_intake_price` · `no_intake_candle`), `sell_quantity_mismatch`, `no_engine_sell`.
3. 오늘 닫힌 외화 성과가 있는데 신선한 환율이 없음(D5 (b) 의 읽기 실패 · 60초 초과) — 그 시도의 진입.
4. 성과 금액 · `closed_at` 이 해석 불가 — 그 시장일의 진입 전부(원장 값 결함, 사람 조사).
5. 시장 시간대 로드 실패 — 그 시장의 진입 전부.
6. (U2 를 Q2 에서 fail-closed 로 고르면) 오늘 부분 청산이 있은 날의 진입 전부.

**거부하지 않아 통과하는 것**(반대 방향도 센다): U1 외부 수동 거래 손실 · U2(허용을 고르면) · U4 미감지 체결 · U5 미실현 손실.
이들은 측정값을 올리지 않으므로 그만큼 한도가 느슨하다.

## 사람 결정 (Q1~Q8) — 권장 포함

| Q | 질문 | 선택지 | 권장 |
|---|---|---|---|
| Q1 | 실현 손익의 원천 | 원장 `trade_outcomes` / 브로커(공식 없음 · WTS 는 엔진 official-only 위반) | 원장 + riskcalc `AsOf` 주석 정정(D2) |
| Q2 | 열린 포지션의 부분 청산 손실(U2) | 제외(명시 한계) / fail-closed / 새 공식(평균가 × 수량) | 제외 + 후속 change(새 공식은 trade-analytics 계약 필요) |
| Q3 | 오늘 닫혔는데 성과 행이 없는 관리 포지션(U3, a130 편입 포함) | fail-closed / 제외 / 초기 위험(수량 × (진입−손절))으로 추정 | fail-closed(거부 목록 2 를 사람이 수용) |
| Q4 | 계좌 단위 한도의 일 경계 | 진입 시장 현지일(관례 재사용) / 시장별 각자 날짜 합산 / KST 고정 / **a136 과 같은 경계**(a136 권장 D1-D: `clock.MarketKR.TradingDay(now − 8h)`, KST 08:00 에 넘어가는 하루) | 진입 시장 현지일(D4). 단 a136(`ea88bb7a`)이 같은 계좌 단위 「하루」를 KST 08:00 경계로 권장했다 — 두 계좌 단위 일일 한도가 다른 하루를 쓰면 운영자에게 설명이 둘이 된다. **a136 D1 과 한 번에 결정할 것** |
| Q5 | 외화 성과 환산 | (a) 봉인 권한만 / (b) 기존 공식 환율 읽기 / (c) 최근 환율 + haircut | (b) + rate budget 계상 |
| Q6 | 전략 경로의 한도 도달 분류 | 현행(결함 → 레인 잠금 가능) / 계좌 단위 진입 차단 + ENTRY_BLOCKED 강화 | 후자 |
| Q7 | 자동 강화의 지속 | 사람 해제까지(현 스펙) / 거래일 경계에서 자동 소멸(스펙 변경 · 완화 방향) | 현 스펙 유지 |
| Q8 | 미실현 손실 포함 | 제외(riskcalc 계약) / 포함(후속) | 제외, 후속 change |

## 구현 순서 (결정 뒤)

1. a136 착지 확인 → base 고정(`capture_change_base.py`) — a136 과 `riskguardian.go` · Guardian policy · risk-management spec 이 겹친다.
2. FLM/BTM: `RiskGuardian.IssueEntry` · `RiskGuardian.PrecheckQFinalEntry` · `engine.riskPolicyFromAutomationGate`(주입 시) ·
   Q6 이면 `strategy_first_leg_admission` 의 거절 분류 함수 + 감독자 하류(F6 미고정 부분).
3. Pre-Edit 선언 → RED(측정 0 이 아닌 날 진입 거절 · 서명값보다 낮추지 않음 · 측정 실패 = daily_loss 단 INPUT_UNAVAILABLE ·
   매도는 측정기 미호출 · 전략 경로 강화) → GREEN → 변이 원장 → 적대적 독립 리뷰.
