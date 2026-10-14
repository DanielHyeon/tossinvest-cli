# a137 — 일일 손실은 측정된 값이다 (daily-loss-is-measured)

## Why

Guardian 판정 체인의 `daily_loss` 단(`risk.checkDailyLoss`)은 있지만, 그 단이 비교하는 **오늘의 실현 손실을 만드는 생산 코드가 없다.**
그래서 `MaxDailyLoss`(절대액) · `MaxDailyLossRatio`(자본비) 한도는 설정 · 인터록 · 콘솔에는 있으나, 실제로는 호출자가 넘긴 숫자를
비교할 뿐이고 그 숫자는 0 이거나 저장소 밖 서명자가 적은 값이다. 사람이 보는 「일일 손실 한도 10만 원」은 지금 보호가 아니다.

### 지금 Guardian 이 일일 손실로 하는 일 (HEAD `7b073f55` · AST · CodeGraph 1.6.0)

분기 근거는 손으로 읽지 않고 `analysis/ast-evidence/*.ast.json`(tools/logic-map 산출물)의 열거를 쓴다. 호출 관계는
`analysis/code-context/raw/`(codegraph 1.6.0) 원문이다.

1. **판정은 있다 — 생산은 없다.** `riskcalc.DailyLoss`(aggregate.go:210-257) 의 호출자는 6 이고 전부 `riskcalc_test.go` 다
   (`raw/callers-DailyLoss.txt`). 생산 호출자 0.
2. **체인 단의 실제 동작** — `risk.checkDailyLoss`(chain.go:518-566, AST 분기 9 · return 10, `risk--checkdailyloss.ast.json`):
   - B1·B2(521·525): 자본 값이 없거나 통화가 다르거나 해석 불가 → `INPUT_UNAVAILABLE` 거절.
   - B3(528): 자본 ≤ 0 → `DAILY_LOSS_LIMIT_REACHED`.
   - B4(535): `Account.DailyRealizedLoss` 를 `magnitudeIn` 으로 읽음 — 비었거나 · 통화 불일치 · **음수**면 `INPUT_UNAVAILABLE`
     (`risk--magnitudein.ast.json` B1~B3: moneyIn 실패 · 해석 실패 · `Sign()<0`).
   - B5·B6(540·543): `riskcalc.WithinLimit` — 한도 이상(근사 동률 포함)이면 `DAILY_LOSS_LIMIT_REACHED`.
   - B7~B9(550·554·559): 손실 ≥ 비율 × 자본이면 `DAILY_LOSS_LIMIT_REACHED`.
   - 결론: **값이 없으면 막는다(fail-closed), 값이 0 이면 통과시킨다.** 값을 스스로 만들지 않는다.
3. **매도는 이 단에 오지 않는다** — `risk.Evaluate`(chain.go:131-144, `risk--evaluate.ast.json`) B2(135) `Side == SideSell` 이면
   `evaluateReduction` 으로 가고 `entryChain`(daily_loss 포함)을 돌지 않는다. 손절 · 비상 청산은 이 한도와 무관하다(지금도, 이 change 뒤에도).
4. **값을 넣는 생산 자리는 둘이고 둘 다 측정이 아니다** (`grep "DailyRealizedLoss" internal cmd` 비시험 + CodeGraph):
   - `engine.(*Tracer).accountState`(tracer.go:467-480, 분기 1 = 시장 통화 선택뿐): 리터럴 `"0"`. 근거는 `requireFlatAccount`(보유 0) 이지만
     **보유 0 은 오늘 손실 0 이 아니다.** 또한 `NewTracer` 의 생산 호출자는 0 이다(`grep -rnF "NewTracer(" cmd internal` 비시험 0).
   - `strategyaccount.sealAuthority`(production.go:200-241, `strategyaccount--sealauthority.ast.json` B1): 서명된 계좌 매니페스트
     (`strategy-account-snapshot-<KR|US>.json`, Ed25519, 신선 창 5초)의 `daily_realized_loss_base` 를 **음수 아님만 확인하고 그대로** 옮긴다.
     이 매니페스트를 만드는 코드는 저장소에 없다(`grep -rln "daily_realized_loss_base"` → 읽는 쪽과 그 시험뿐). 생산 전략 진입
     (`strategy_first_leg_admission.go:91` → `PrecheckQFinalCampaignFirstLeg` → `PrecheckQFinalEntry`)이 이 값을 체인에 넣는다.
5. **한도 도달의 자동 강화는 한 경로에만 있다** — `escalateFor`(riskguardian.go:640-656) 의 호출자는 `IssueEntry` 하나다
   (`raw/callers-escalateFor.txt`). 생산 전략 경로인 `PrecheckQFinalEntry`(riskguardian_qfinal.go:82-167, AST 분기 17) 의 B16(158)
   체인 거절은 `chainRefusal` 만 돌려주고 운영 모드를 올리지 않는다. 그 거절은 호출자(strategy_first_leg_admission.go:93)에서
   범위 거절이 아닌 `AUTHORITY_MISMATCH` 결함으로 분류된다.
6. `policy.RiskBudget = dailyLoss`(guardian_wiring.go:59)는 **건당 위험 예산**이다(risk-management 「Automation gate limits are the
   production Guardian policy source」 — "per-trade risk budget SHALL equal the configured maximum daily-loss amount"). 일일 손실 측정과
   다른 양이며 이 change 의 대상이 아니다.

## What Changes (권장안 — 사람 결정 Q1~Q8 뒤 확정, design.md)

1. **원장 기반 일일 손실 측정기** — 엔진 원장(`journal` 패키지만 경유)의 `trade_outcomes.realized_pnl_after_costs`(비용 차감 후,
   부호 있음)를 진입 시장의 현지 거래일(`clock.Market.TradingDay` — 일일 손실 예약 소멸과 같은 경계)로 모아
   `riskcalc.DailyLoss` 로 계산한다. 공식은 riskcalc 의 기존 계약(손실 = max(0, −순실현), 미실현 제외)을 그대로 쓴다.
2. **Guardian 이 측정값을 바닥으로 쓴다** — 진입 판정 직전에 `effective = max(호출자 값, 측정값)`. 호출자(서명 매니페스트 · tracer)가
   넘긴 값을 낮추지 않는다. 측정 실패는 `daily_loss` 단의 `INPUT_UNAVAILABLE` 로 **진입만** 거절한다(체인 순서 보존).
3. **매도 · 손절 · 비상 청산 경로는 측정기를 부르지 않는다** — `IssueReduction` · `evaluateReduction` 무편집, 시험으로 고정.
4. (Q6) 생산 전략 경로에서 한도 도달을 결함이 아닌 정책 결과로 분류하고, `IssueEntry` 와 같은 자동 강화(ENTRY_BLOCKED)를 건다.

## What Does NOT Change

- 체인 순서 · 단 이름 · reason code 집합 · `checkDailyLoss` 본문 · `riskcalc.DailyLoss` 공식 · 한도 값 · 설정 스키마 · 원장 스키마.
- 매도 · 손절 · 익절 · 비상 청산 · 대사 · 체결 감지 경로. 새 브로커 전송 0(Q5 에서 기존 환율 읽기 재사용 여부만 결정).
- 미실현 손익 포함(riskcalc 계약이 정의하지 않음 — 후속 change, 보수 방향만).
- 다중 horizon 진입 손실 잠금(`ActivateEntryLossLock` 도 생산 호출자 0 — 인접 결손, 별도 change).

## Impact

- Affected specs: `risk-management`(요구 1개 추가 — 「일일 손실 한도의 입력은 원장에서 측정된다」).
- Affected code(예정): `internal/execgw/riskguardian.go`(`IssueEntry`) · `riskguardian_qfinal.go`(`PrecheckQFinalEntry`) ·
  `internal/app/engine/guardian_wiring.go`(측정기 주입) · 새 파일(측정기) · 필요 시 `strategy_first_leg_admission.go`(Q6).
  **병행 a136(Guardian 일일 매수 notional 한도)과 같은 파일 표면** — a136 착지 뒤에 base 를 고정한다.
- 위험 등급: **High-risk**(Guardian · 위험 입력). FLM/BTM 선행 · Pre-Edit 선언 · 변이 원장 · 적대적 독립 리뷰 의무.
- 상태: **사람 결정 대기로 proposal/design 에서 멈춤** — 측정 원천(Q1), 부분 청산(Q2), 성과 동결 거절 포지션(Q3), 시장 간 일 경계(Q4),
  환율(Q5), 전략 경로 분류(Q6), 강화의 지속(Q7), 미실현(Q8). design.md 「사람 결정」.
