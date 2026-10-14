# a137 design — 일일 손실은 측정된 값이다

> 상태(2026-10-14): 사용자 결정 Q1~Q8 · Q12 · Q13 과 Q5 FX 조항 승인을 반영했고(원문은 review.md §2), proposal-freeze 리뷰
> **APPROVE-WITH-FIXES**(P1×5 · P2×6 · P3×4, review.md §4)의 수리도 반영했다. 구현은 a136 이 공유 production 파일
> (`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는 risk-management 요구)을 커밋한 뒤 시작한다.
> 그 전에는 이 change 의 새 파일만 만든다. base 는 이 change 의 첫 Go 커밋의 부모로 고정한다. archive 는 a136 뒤에 한다.

## D0. 현재 동작 — 증거

분기 주장은 `analysis/ast-evidence/`(tools/logic-map, HEAD `7b073f55`)와 `analysis/function-logic/`(편집 전 좌표)의 열거에서 왔다. 호출 관계는
codegraph 1.6.0(`analysis/code-context/raw/`)이다. 「정정」 행은 두 번의 독립 리뷰로 고친 것이다(review.md §1 · §4).

| # | 사실 | 근거 |
|---|---|---|
| F1 | `checkDailyLoss` 는 입력 부재 · 통화 불일치 · 음수 · 해석 불가 · 한도/비율 계산 오류를 `INPUT_UNAVAILABLE` 로 거절한다(B1 · B2 · B4 · B5 · B7 · B8). `DAILY_LOSS_LIMIT_REACHED` 는 자본 ≤ 0(B3) · 절대액 이상(B6) · 비율 이상(B9) 셋뿐이다. 0 은 통과 | `risk--checkdailyloss.ast.json` · `risk--magnitudein.ast.json` |
| F2 | 매도 의도는 `entryChain` 에 들어가지 않는다 | `risk--evaluate.ast.json` B2(135) |
| F3 | `riskcalc.DailyLoss` 생산 호출자 0. 결과는 `max(0, −순실현)` 이라 성과 0 건은 0(B10)이다 | `raw/callers-DailyLoss.txt` · `riskcalc--dailyloss.ast.json` B10 |
| F4 | `riskcalc.DailyLoss` 는 모든 항목의 `TradingDay` 가 같고 · `CostsDeducted` 이고 · `AsOf` 가 10초 신선 한계 안이어야 하며, `checkFresh` 는 미래 `AsOf` 도 거절한다 | `riskcalc--dailyloss.ast.json` B4~B9 · riskcalc.go:287-302 |
| F5 | 체인에 값을 넣는 생산 자리: tracer 리터럴 `"0"`(생산 생성자 0) · 서명 계좌 매니페스트 `daily_realized_loss_base`(저장소 안 생산자 0) · 분석용 `measure/counterfactual.go:335` `"0"`(주문 경로 아님) | grep · AST |
| F6 | 살아 있는 진입 입구는 `PrecheckQFinalEntry` 하나다. `IssueEntry` 의 생산 경로 둘(`Tracer` · `GuardianAdapter`)은 생산 생성 0 이라 `escalateFor` 는 잠들어 있다. 체인 없는 `execgw.Issuer.IssueEntry`(issue.go:106)도 생산 호출자 0 | callers · grep |
| F7 | **정정(2차).** 전략 경로 한도 거절의 하류는 레인 종류에 따라 갈린다. `runMarket` 의 `refreshOnly = !effective && RefreshesAuthority`(:907). 서명 활성화가 없는 레인(`effective` 거짓)은 B12 에서 `recordSwallowedCycleError` + `continue` — 폴링마다 재거절, 강화 없음. **서명 활성화된 레인**(`Effective: true`, :511)은 `refreshOnly` 거짓이라 (966) `latchMarket` 로 **durable 레인 잠금**이 된다. ENTRY_BLOCKED 뒤에는 `dispatch` 의 `ObserveStrategyEntryGate`(strategy_dispatch_cycle.go:143)가 오류로 거절해 다른 레인도 같은 길로 잠긴다 | `function-logic/…runmarket` B12 · (966) · 리뷰 P1-3 |
| F8 | 원인 상실: admission.go:93 이 `err` 를 `cause` 로 싣지 않고 dispatch_cycle.go:157 은 `Detail` 만 포맷한다 | `function-logic/…admit` B5 · `…dispatch` B12 · B13 |
| F9 | 성과 행은 포지션이 완전히 닫힐 때 동결되고, 금액은 인스턴스 수명 전체(부분 익절 포함)의 순실현이다. 동결 거절 20 종은 원장에 남지 않는다. a130 backfill 은 체결로 닫힌 편입 포지션만 되메운다 | trade_outcomes.go · trade_outcome_refusal.go |
| F10 | `positions.closed_at` 은 **원장 커밋 시각**이다(체결 반영 트랜잭션의 `fill.CommittedAt`, position_projection.go:316-318 · fills.go:174 / 조정은 조정 트랜잭션의 now). 브로커 체결 시각이 아니다 | 리뷰 P1-1 |
| F11 | 공식 Open API 에 실현 손익 조회가 없고 엔진은 official-only 다 | `internal/official` endpoint 열거 |
| F12 | 원장은 단일 연결(`SetMaxOpenConns(1)`, journal.go:174)이다. `PrecheckQFinal*` 은 ctx 를 받지 않고 원장 변경 · 브로커 호출이 없는 계약이다(riskguardian_qfinal.go:79-81) | HEAD |
| F13 | 조정이 수량을 0 으로 만들면 같은 트랜잭션이 `state = CLOSED` · `closed_at = now` 로 바꾸고 `position_adjustments.new_quantity` 0 행이 남는다 | position_adjustments.go:332-352 · `convergedState` |
| F14 | 게이트 OFF 이면 생산 Guardian 이 만들어지지 않는다(engine.go:536 `gate.Enabled && guardian == nil`) | HEAD |

## D1. 측정 모집단과 산식

일 키 `D` 와 귀속 함수는 a136 의 `clock.AccountTradingDay`(`= clock.MarketKR.TradingDay(t − 8h)`, KST 08:00 롤오버) **하나**다. a137 은 그것을
부르고 복제하지 않는다.

1. **귀속(정정 P1-1)**: 계좌의 `positions` 중 `state = CLOSED` 이고 `clock.AccountTradingDay(closed_at) == D`, 즉 `closed_at ∈ [D 08:00 KST, D+1 08:00 KST)`.
   시장별 `TradingDay` 로 귀속하지 않는다 — `closed_at` 은 원장 커밋 시각(F10)이라, 시장 현지일로 귀속하면 늦게 기록된 손실이 이미 끝난 날로 떨어져
   **영원히 세어지지 않는다**(fail-open). 계좌 일로 귀속하면 늦은 기록은 언제나 지금의 일 키에 들어간다. 이 정의는 사용자 Q4 의 「KST 08:00 롤오버 하루」
   문자 그대로다: D 의 하루에는 서울 D 의 KR 정규장과 뉴욕 D 의 US 정규장이 들어간다.
2. **시장**: KR · US 만 측정 대상이다. 그 밖의 시장 값 · 해석 불가 시장은 측정 불가(P3-2 — 하나로 통일).
3. **제외**: 관리 대상이 아닌 포지션(`position.ExitEligible(entry_decision_id, adoption_id)` 거짓 — 외부 · 수동 보유).
4. **조정 종결(Q13, 정정 P1-2)**: 마지막 조정의 `new_quantity` 가 **decimal 비교로** 0 인 포지션은 사람이 앱에서 판 것으로 보아 세지도 막지도 않는다.
   단 그 인스턴스의 수명(`opened_at` 이후) 안에 **엔진 매도가 브로커에 나갔을 수 있으면** 측정 불가(fail-closed). Q13 의 제외는 사람의 앱 매도에만 준다.
   「엔진 매도가 나갔을 수 있음」은 둘 중 하나다(재확인 P2 — 보호 주문 포함):
   - 같은 계좌 · 시장 · 심볼의 엔진 SELL **PLACE** attempt 로 상태가 NOT_DISPATCHED · FAILED_CONFIRMED 가 아닌 것(취소 · 정정 attempt 는 그 자체로 매도가 아니므로 세지 않음).
   - 같은 범위의 브로커 대기 보호 주문: 인스턴스 수명과 **겹치는** `protection_sagas` · `protection_mutation_attempts`(수명 안 생성만이 아님 — §4.2 P3-3) 중 kind 가 **CREATE · REPLACE** 이고 state 가
     PLANNED 가 아닌 것(DISPATCHED · ACKNOWLEDGED · IN_DOUBT · CLOSED — 브로커에 놓였을 수 있는 보호 매도). CANCEL 은 세지 않는다.
5. 남은 각 포지션: 성과 행이 있으면 `realized_pnl_after_costs`, 없으면 **측정 불가**(Q3).
6. **Q12(정정 P2-3 — 범위 축소)**: 포지션 **사이에서** 이익은 손실을 상쇄하지 않는다 — 음수 행만 남긴다. 한 포지션 수명 안의 순실현(부분 익절 이익이
   마지막 손실을 줄이는 것)은 성과 행 그대로다(U8). 그래서 「하루 안에서 측정 합계는 줄지 않는다」는 포지션 단위로만 참이다.
7. **환산(D5)**: KR → KRW, US → USD 는 포지션 시장에서 유도한다(성과 행에 통화 칸이 없다). 기준 통화와 같은 통화의 손실은 환산하지 않는다.
   다른 통화는 그 시장의 봉인 권한으로 환산하고 **올림**(ceil — 손실을 작게 만들지 않음, P3-1)한다.
8. `riskcalc.DailyLoss{TradingDay: D, TargetCurrency: base, Realized: [{TradingDay: D, Amount: −Σ손실, Currency: base, CostsDeducted: true, AsOf: 측정 순간}]}`.
   공식은 riskcalc 그대로다. `AsOf ≤` 체인 `Now`(F4).

## D2. 측정의 자리 — Guardian, 체인 직전, 한 번

- **측정 토큰.** 원장 읽기는 precheck 앞에서 Guardian 의 새 메서드 `MeasureDailyLoss(ctx, D)` 가 한다. 결과는 다른 패키지가 만들 수 없는 불투명 토큰이다
  (내보내지 않는 칸만 — `QFinalEntryPrecheck` 와 같은 패턴). 토큰은 통화별 손실 합(환산 전) · 일 키 · 측정 순간 · 발급 Guardian 신원 · **측정 실패면 봉인된 오류**를 담는다.
- **측정 실패의 운반(P2-2).** 측정 실패는 권한 수집 오류(`AUTHORITY_COLLECTION_FAILED`)로 올리지 않는다 — 그러면 kill switch 같은 앞 단의 사유가 가려진다.
  실패도 토큰 안에 봉인된 오류로 실려 precheck 까지 가고, 거기서 `daily_loss` 단의 `INPUT_UNAVAILABLE` 이 된다. 첫 다리 조립은 `MeasureDailyLoss` 를
  **마지막 단계**로 부른다(측정 순간이 precheck 에 가장 가깝게).
- **precheck 의 확인.** ① 토큰이 이 Guardian 의 것 ② 측정 순간이 Guardian 시계 기준 10초 안이고 미래가 아님 ③ 일 키가 지금의 `AccountTradingDay(now)` 와 같음
  ④ 봉인 오류 없음. 하나라도 어긋나거나 토큰이 영값이면 측정 실패.
- **바닥 규칙(P1-4).** `effective = max(호출자 값, 측정값)` 은 **둘 다 유효할 때만**이다.
  - 호출자 값이 무효(해석 불가 · 통화 불일치 · 음수)면 바꾸지 않고 그대로 둔다 — 체인이 지금처럼 `INPUT_UNAVAILABLE` 로 거절한다(측정값으로 대체 금지).
  - 측정 실패면 호출자 값으로 **떨어지지 않는다** — 쓸 수 없는 값으로 두어 `daily_loss` 단이 거절하게 하고, 그 `Detail` 을 측정 실패 문구로 바꾼다.
  - 앞 단(kill switch · 운영 모드 · latch …)이 먼저 거절하면 그 사유가 그대로다.
- **Detail(P2-6).** 측정 실패 문구는 시장 · 심볼 · 포지션 id 해시(`pos-` 접두 — 원장 id 원문 아님) · 「성과 행 없음」/「엔진 매도가 섞인 조정 종결」/「환산 권한 없음」 등
  사유 하나를 싣는다. 계좌 참조 · 원문 오류 · 금액은 싣지 않는다(a131/a133 경계).
- precheck 에서 **한 번** 잰다. `IssuePrecheckedQFinal*` 은 다시 재지 않는다 — precheck 와 발급 사이에 닫힌 손실은 그 한 건이 못 보고 다음 진입이 본다(U7).
- **원장 읽기(재확인 P2 — 단일 연결).** 계좌 · `state = CLOSED` 로 한정하고 `closed_at` 범위(일 키 창 ± 작은 여유, RFC3339 UTC 문자열 비교)로 미리 거른
  색인 읽기 한 번이다. 문자열 범위는 **사전 필터일 뿐**이고 귀속은 Go 에서 `closed_at` 을 해석해 정한다. 사전 필터가 형식이 어긋난 행을 떨어뜨리지 않도록
  **같은 질의 안에서** 계좌의 CLOSED 관리 행 중 `closed_at` 이 **`Z` 로 끝나는 정규 RFC3339 UTC 가 아닌 것**(NULL · 빈 값 포함 — 사전 필터와 같은 정의, §4.2 P3-2)의
  존재를 따로 세고, 하나라도 있으면 측정 불가로 한다.
  행을 연 채 환산 · 네트워크를 하지 않는다(F12). 환산은 봉인된 권한의 산술뿐이라 브로커 호출이 없다. 현실 크기 픽스처에서 지연을 잰다(tasks 2.4).
- `IssueEntry`(잠든 입구)에도 같은 바닥을 건다 — 나중에 깨어나도 0 을 믿지 않게. 체인 없는 `execgw.Issuer.IssueEntry` 는 생산 호출자 0 을 시험으로 고정한다.
- `IssueReduction` · `evaluateReduction` · 청산 · 손절 경로는 측정기를 부르지 않는다(시험으로 고정).
- 게이트 OFF: Guardian 이 없으므로 측정도 없다(F14 — 토글 OFF = upstream 동작).

## D3. 전략 경로의 분류와 강화 (Q6, 정정 P1-3)

- 강화는 첫 다리 승인(`admit`)이 한다 — precheck 는 원장을 바꾸지 않는 계약이다. precheck 거절에서 `errors.As(err, *execgw.IssueRefusal)` 이고
  **`Stage == StageChain` 이고 `Reason == DAILY_LOSS_LIMIT_REACHED`** 일 때만(P2-1, 타입 · 코드로) Guardian 의 새 강화 입구(ctx, ENTRY_BLOCKED 영속)를 부르고
  새 코드 `DAILY_LOSS_LIMIT` 을 돌려준다. cause = 그 거절(강화 오류는 `errors.Join`). 다른 precheck 거절도 cause 를 싣는다(F8).
- **주기 끝내기 — 좁게(재확인 판정 BLOCKED 수리, Manager 결정).** `runMarket` 의 `if err == nil { continue }`(:935)는 삼킴(:938)과 잠금(:966)보다 앞이므로
  nil 주기 반환이 둘 다 피한다. 그래서 `dispatch` 는 아래 셋**만** 오류가 아닌 「계좌 진입 차단」 결과로 돌려준다:
  1. 첫 다리 승인의 `DAILY_LOSS_LIMIT`(D3 첫 항목).
  2. `ObserveStrategyEntryGate` 의 운영 모드 투영 거절 중 **차단 중인 모드 행의 trigger 가 `ModeTriggerDailyLossLimit` 이고 `Actor == AUTO`** 인 것(§4.2 P3-1 —
     `Cause` 는 문자열이라 사람 전이의 사유 문구가 trigger 철자와 같아도 조용한 경로에 들지 않게). 오늘의 거절은
     `ReasonOperatingModeBlocked` 하나(modegate.go:61)이고 기동 복원 실패 latch(mode_projection_wiring.go:33)도 같은 코드이며 `RejectedError` 는 Reason/Detail 뿐이다.
     그래서 모드 투영이 trigger 를 **타입으로** 실어야 한다(새 칸 — 문구로 가르지 않음). trigger 를 모르거나 다른 trigger(다른 ENTRY_BLOCKED · HALT_ALL 사유 ·
     기동 복원 실패)면 **오늘과 바이트 단위로 같은 동작**(오류 → effective 레인은 durable 잠금)이다.
  3. 경합: 레인 B 가 강화 **전에** 승인되어 Gateway 의 `withStrategyEntryGateAuthority` 재확인(strategy_entry_gate_authority.go:62-71)에서 세대/모드가 바뀌어
     `ReasonStrategyDispatchFenced`(gateway.go:712-716)로 거절될 때, 그 변경의 원인이 trigger `ModeTriggerDailyLossLimit` 인 모드 변경이면 같은 「계좌 진입 차단」 결과.
     그 밖의 원인(심볼 latch · 다른 모드 trigger · 세대 변화만)은 오늘과 같다.
  (P2-2) 일일 손실 모드 latch 가 서 있는 동안에는 모드가 latchOrder 맨 끝이어도 조회 신선도(query-stale) 거절은 가려진다 — latch 가
  stale 판정보다 먼저 읽히기 때문이며 의도된 것이다(어차피 진입은 막혀 있다). 모드가 풀리면 stale 은 다음 관측에서 다시 보인다.
  전달 루프는 그 결과를 보면 그 주기의 나머지 범위 전달을 멈춘다. **앞서 건너뛴 범위 거절의 보고는 오늘과 같다**: 같은 주기에 이미 `*strategyScopeRefusal`
  (예: 버킷 고갈, admission.go:99-101)로 건너뛴 범위가 있으면 `deliverEachStrategyHandoff` 는 지금처럼 `errors.Join(skipped...)` 를 돌려준다(J4 — 조용한 건너뛰기 금지).
  그 경우 주기는 오류로 끝나고 effective 레인은 오늘과 같은 길을 간다 — 이것은 명시된 · 시험되는 동작이다(혼합 경우). `runMarket` 은 편집하지 않는다.
- 분류 불변(P2-1): 모든 precheck 거절과 강화의 `errors.Join` 에 대해 `isCentralStrategyIntegrity` · `*strategyScopeRefusal` 판정이 편집 전과 같음을 표 시험으로 고정한다.
- **측정 불가(구현 리뷰 P1-3, Manager 결정 ①)**: 측정기 · 토큰에서 생긴 daily_loss 단 INPUT_UNAVAILABLE 은 `*execgw.DailyLossUnmeasurable`
  표지를 원인으로 실어(`dailyLossChainRefusal`) 첫 다리 승인이 `DAILY_LOSS_UNMEASURABLE` 로 분류하고, 전달 루프는 계좌 차단과 같은 자리에서 주기를
  조용히 끝낸다. 강화는 없고, 진입은 잴 수 있을 때까지 매 주기 다시 거절된다(fail-closed 유지). 표지가 없는 입력 불가는 오늘처럼 결함(서명 활성화
  레인은 잠김). 두 레인 종류 모두 시험으로 고정.
- **측정 불가의 흔적(재리뷰 P1-N1)**: 조용한 종료는 레인 잠금 · 삼킴 계수 · 운영 모드를 건드리지 않으므로, 감독자 스냅숏에 시장별
  `DailyLossUnmeasurableCycles`(조용히 끝난 측정 불가 주기 수) · `FirstDailyLossUnmeasurable`(첫 원인 — dailyloss 고정 문구 + 시장 · 심볼 · `pos-`
  손잡이, 계좌 값 · 원문 오류 없음)을 두고, 경고 줄 `engine.daily_loss_unmeasurable` 를 **계좌 일 키마다 한 번** 낸다. 기록은 주기 ctx 에 실린 기록자를
  `dispatchStrategyMarketHandoffs` 가 부르는 방식(전달 루프 반환값 불변).
- **원인 판정(재리뷰 P2-N1)**: daily_loss 단은 자본도 해석하므로, 측정 실패로 그 단이 입력 불가면 손실만 0(기준 통화)으로 바꿔 체인을 한 번 더 돌린다.
  여전히 입력 불가면 원인은 손실이 아니다 — 상세를 바꾸지 않고 측정 불가 표지도 달지 않는다(오늘처럼 결함, 서명 활성화 레인은 잠김).
- **자본 ≤ 0(구현 리뷰 P2-3)**: 체인 B3 은 자본이 0 이하면 `DAILY_LOSS_LIMIT_REACHED` 로 거절한다(기존). 살아 있는 문에서도 같은 사유이므로 첫 다리
  승인이 ENTRY_BLOCKED 로 강화한다 — 보수 방향이고 그대로 둔다.
- **강화 통지 실패**: `escalateFor` 는 전이가 커밋됐어도 통지 기록이 실패하면(`ErrModeAnnouncementFailed`) 오류를 돌려준다. 그 경우 첫 다리 승인은
  계좌 차단 표지 대신 거절 + 오류를 join 한 결함으로 남긴다(시끄러운 쪽) — 모드는 이미 ENTRY_BLOCKED 라 다음 주기부터는 모드 차단 표지로 조용해진다.
- Q7: ENTRY_BLOCKED 는 사람 해제까지 유지된다. **재강화 함정(P2-6)**: 같은 일 키 안에서 사람이 풀어도 측정 합계는 줄지 않으므로(D1-6) 다음 진입이 곧바로 다시
  강화한다. 같은 날 다시 진입하려면 한도를 올리거나(완화 — 사람 승인 · audit) 일 경계를 기다려야 한다. 운영 문서에 적는다.

## D4. 원장이 세지 못하는 손실 (명시 한계)

| # | 빠지는 손실 | 결정 |
|---|---|---|
| U1 | 외부 수동 거래 · 관리 대상이 아닌 포지션 | 세지 않음(F11) |
| U2 | 열린 포지션의 부분 청산 | Q2 제외. 그 포지션이 완전히 닫히는 날 성과 행이 수명 전체를 담아 그날에 들어간다(F9) |
| U3 | 사람의 앱 매도로 조정 종결된 관리 포지션(엔진 SELL 발송 없음) | Q13 세지 않음 · 막지 않음 |
| U4 | 브로커는 체결했으나 원장 미반영 | 기존 「체결 감지 5주기 실패 → ENTRY_BLOCKED」가 맡음. 늦게 반영되면 그 순간의 일 키에 들어감(D1-1) |
| U5 | 미실현 | Q8 제외 |
| U7 | precheck 뒤 · 발급 전에 닫힌 손실 | 그 한 건은 못 봄, 다음 진입이 봄 |
| U8 | 한 포지션 수명 안의 이익이 같은 포지션의 손실을 줄임(부분 익절 뒤 손절) | 성과 행 그대로(Q12 는 포지션 사이에만) |

## D5. 환율 — 승인된 FX 조항 (사용자 승인 2026-10-14, 선택 「1」)

- 근거 문장: risk-management 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」 — 요청은 자기 request-scoped frozen FX authority 를 쓰고 「다른 환율 사용 금지」.
- 승인된 조항(델타의 MODIFIED, 원문은 review.md §2): 기준 통화 손실은 환산 · 권한이 필요 없다. 다른 시장의 손실(그리고 a136 의 US 수동 매수 금액)은 **같은 paired FX
  collection(같은 `observedAt`)에 실린 그 시장의 권한**으로만 환산하고, Guardian 이 precheck 에서 공식 원천 · 통화쌍 · 신선도를 검증한다. 없으면 진입 fail-closed.
  재조회 · 호출자 환율 · 다른 파도의 권한은 금지. 요청의 sizing · exposure · reservation · envelope FX 규칙은 그대로.
- 운반: 첫 다리 조립이 같은 파도의 짝 FX(`strategyFXAuthorityPair` — 같은 `observedAt`)에서 다른 시장 권한을 꺼내 `QFinalEntryIssuance` 에 싣는다. US 권한은 US 후보가
  Ready 일 때만 봉인된다(strategy_fx_authority.go:205) — 없으면 그 KR 진입은 측정 불가.
- 빈도(재확인 P2 로 정정): US 정규장 손실은 대개 KST 08:00 **전**에 원장에 반영되어 전 일 키에 속하므로 보통 KR 진입은 USD 환산을 요구하지 않는다. 그러나
  US 손실이 KST 08:00 **뒤**에 반영되면(05:00~09:00 배포 · 재시작 창에서 체결 감지가 늦게 따라잡는 경우, 또는 US 확장 시간대 청산) 그 손실은 오늘 일 키에 들어가고,
  KR 정규장 동안 US 권한은 봉인되지 않으므로 **그날 KR 진입 전부가 fail-closed** 된다(거부 목록 (e)). 사용자가 fail-closed 를 골랐으므로 동작은 유지한다(관찰, 측정 아님).
- **조항의 자리(결정됨)**: 조항은 a136 델타의 risk-management MODIFIED 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」에 있다(a137 문구 + 수동 매수 시나리오 하나). a137 델타는 그 MODIFIED 를 두지 않고 인용만 한다 — 같은 요구를 두 델타가 MODIFIED 하면 뒤의 archive 가 앞을 덮는다. 구현은 a136 의 `PairedFXAuthority` · `qFinalManualBuyFX` 를 그대로 재사용한다.

## Fail-closed 가 거부하는 정상 입력 (진입만, P2-5)

매도 · 손절 · 비상 청산 · 대사 · 체결 감지는 측정기를 부르지 않으므로 어느 것도 거부되지 않는다. 빈도는 추정이고 측정이 아니다.

| # | 거부 | 범위 · 지속 | 빈도 추정 |
|---|---|---|---|
| (a) | 원장 읽기 오류(busy · 잠금 · 닫힌 핸들) | 그 시도 한 번 | 드묾(단일 연결 경합 시) |
| (b) | 그날 체결로 닫힌 관리 포지션의 성과 행 부재 — 편입 포지션 | 그 일 키 전체, a130 backfill 로 해제 | 편입 포지션을 엔진이 판 날의 일부(`needs_intake_price` 등) |
| (c) | 같은 부재 — 엔진 진입 포지션(`cost_estimate_failed` · `insert_failed` · `round_trip_unpriceable` · `sell_unpriced` · `sell_quantity_mismatch` …) | 그 일 키 전체, **다음 일 경계까지 복구 경로 없음** | 미측정 — 아래 질의로 사람이 잴 것 |
| (d) | 엔진 매도(SELL PLACE 발송 — 오래 묵은 IN_DOUBT SELL(settled_at NULL)도 끝이 없으므로 포함 — 또는 브로커 대기 보호 주문 CREATE/REPLACE)가 있던 수명의 조정 종결 | 그 일 키 전체 | 보호 주문이 WIRED 면 드물지 않을 수 있음(관리 포지션 대부분이 보호 주문을 가짐) |
| (e) | KR 진입에 같은 일 키의 US 손실이 있는데 짝 FX 의 US 권한 없음 | **그 KR 세션 내내**(KR 정규장 동안 US 후보가 Ready 가 아니라 US 권한이 봉인되지 않음) | 드물지 않을 수 있음: US 손실이 KST 08:00 **뒤**에 원장에 반영되면(05:00~09:00 배포 · 재시작 창의 늦은 체결 반영) 그날 KR 진입 전부가 막힘. 동작은 유지(사용자 fail-closed 선택) |
| (f) | `closed_at`(NULL · 빈 값 · 비 RFC3339 포함 — 사전 필터에 걸리지 않아도 셈) · 성과 금액 · 시장 해석 불가 | 그 일 키 전체(원장 값 결함, 사람 조사) — 형식 결함 행은 일 키와 무관하게 **매일** 막음 | 매우 드묾 |
| (g) | 측정 토큰 부재 · 만료(10초) · 미래 · 일 키 불일치 · 다른 Guardian | 그 시도 | 조립 결함 또는 KST 08:00 경계를 걸친 precheck |

**사람이 잴 읽기 전용 질의**(에이전트는 실 원장에 돌리지 않는다 — 복사본에 `sqlite3 -readonly`):

```sql
SELECT p.market,
       COUNT(*)                         AS closed_managed,
       SUM(o.position_id IS NULL)       AS without_outcome
  FROM positions p
  LEFT JOIN trade_outcomes o ON o.position_id = p.id
 WHERE p.state = 'CLOSED'
   AND (coalesce(p.entry_decision_id,'') <> '' OR coalesce(p.adoption_id,'') <> '')
   AND NOT EXISTS (SELECT 1 FROM position_adjustments a
                    WHERE a.position_id = p.id AND CAST(a.new_quantity AS REAL) = 0)
 GROUP BY p.market;
```

`without_outcome / closed_managed` 가 (b)+(c) 가 하루를 막는 비율의 상한 추정이다.

거부하지 않아 통과하는 것: U1 · U2 · U3 · U4(반영 전) · U5 · U7 · U8.

## 구현 순서

1. (지금, 새 파일만) 측정기 순수 함수 패키지의 RED(태그 `a137red` — 공유 체크아웃의 `go test ./...` 를 깨지 않음) · 엔진 쪽 편집 대상 FLM/BTM
   (`admit` · `dispatch` · `runMarket`(무편집 근거)) · 리뷰 수리 반영.
2. a136 공유 파일 커밋 확인(`git log -1 -- <file>` · `git diff`) → base = 이 change 첫 Go 커밋의 부모 → riskguardian*.go FLM 재추출 → Pre-Edit → GREEN.
3. 변이 원장 → 적대적 독립 리뷰 → Manager 게이트 → a136 뒤 archive.
