# a137 design — 일일 손실은 측정된 값이다

> 상태(2026-10-14): **사용자 결정 반영(Q1~Q8 · Q12 · Q13, 기록 원문은 review.md)**. 남은 질문 하나 — **Q5 의 스펙 준수**(D5). 구현 순서는
> a136 이 공유 production 파일(`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는 risk-management 요구)을
> 먼저 커밋한 뒤다. 그 전에는 이 change 의 새 파일만 만든다. base 는 이 change 의 첫 Go 커밋의 부모로 고정한다.

## D0. 현재 동작 — 증거 (독립 대조 판정 ACCURATE-WITH-CORRECTIONS 반영)

분기 주장은 `analysis/ast-evidence/`(tools/logic-map, HEAD `7b073f55`)의 열거에서 왔다. 호출 관계는 codegraph 1.6.0(`analysis/code-context/raw/`)이다.
정정된 행에는 「정정」이라고 적었다(리뷰 기록은 review.md).

| # | 사실 | 근거 |
|---|---|---|
| F1 | `checkDailyLoss` 는 입력 부재 · 통화 불일치 · 음수 · 해석 불가 · 한도/비율 계산 오류를 `INPUT_UNAVAILABLE` 로 거절한다(B1 · B2 · B4 · B5 · B7 · B8). `DAILY_LOSS_LIMIT_REACHED` 는 자본 ≤ 0(B3) · 절대액 이상(B6) · 비율 이상(B9) 셋뿐이다. 0 은 통과 — **정정**(B5 · B7 · B8 을 한도 도달로 적었던 것) | `risk--checkdailyloss.ast.json` · `risk--magnitudein.ast.json` |
| F2 | 매도 의도는 `entryChain` 에 들어가지 않는다 | `risk--evaluate.ast.json` B2(135) |
| F3 | `riskcalc.DailyLoss` 생산 호출자 0. 결과는 `max(0, −순실현)` 이라 성과 0 건은 0(B10)으로 통과한다 | `raw/callers-DailyLoss.txt` · `riskcalc--dailyloss.ast.json` B10 |
| F4 | `riskcalc.DailyLoss` 는 모든 항목의 `TradingDay` 가 같고 · `CostsDeducted` 이고 · `AsOf` 가 계좌 스냅샷 신선 한계(10초) 안이어야 한다. `checkFresh` 는 `AsOf` 가 `Now` 보다 미래여도 거절한다 | `riskcalc--dailyloss.ast.json` B4~B9 · riskcalc.go `checkFresh` |
| F5 | 체인에 값을 넣는 생산 자리: tracer 리터럴 `"0"`(생산 생성자 0) · 서명 계좌 매니페스트 `daily_realized_loss_base`(저장소 안 생산자 0) · 분석용 `measure/counterfactual.go:335` `"0"`(주문 경로 아님) | grep · AST |
| F6 | **살아 있는 진입 입구는 `PrecheckQFinalEntry` 하나다 — 정정.** `IssueEntry` 의 생산 경로는 둘 다 죽어 있다: `Tracer.submitEntry`(`NewTracer` 생산 호출자 0) · `IssueStrategyEntry` ← `GuardianAdapter.IssueAndPlan`(`GuardianAdapter` 생산 생성 0). 그래서 `escalateFor` 는 생산에서 **잠들어 있다**. 체인을 돌지 않는 `execgw.Issuer.IssueEntry`(issue.go:106)도 생산 호출자 0 | callers · grep |
| F7 | **전략 경로의 한도 거절 하류 — 정정(레인 잠금 아님).** 생산 감독자는 `NewRefreshingPairedStrategyEntrySupervisor`(cmd/tossctl/engine.go:690, `RefreshesAuthority: true`)라 `refreshOnly` 가 참이다. 경로: `PrecheckQFinalEntry` 체인 거절 → `strategy_first_leg_admission.go:93` `AUTHORITY_MISMATCH`(원인 **버림**) → `strategy_dispatch_cycle.go:157` 이 `Detail` 문자열만 감싼 오류 → 감독자 `:937-960` `recordSwallowedCycleError` + `continue`. `latchMarket`(:966)에 닿지 않고, 그 주기의 뒤 범위는 평가되지 않으며(handoff_delivery.go:68-79 — 범위 거절이 아니므로 첫 오류에서 멈춤), 다음 폴링마다 다시 거절되고 강화는 없다 | HEAD 읽기 · 독립 대조 P1-1 |
| F8 | 원인 상실: :93 이 `err` 를 `cause` 로 싣지 않고 :157 은 `Detail` 만 포맷하므로 하류의 `errors.As` 가 `*IssueRefusal` 을 볼 수 없다 | 독립 대조 P2-2 |
| F9 | 성과 행은 포지션이 **완전히 닫힐 때** 동결되고, 그 금액은 **인스턴스 수명 전체**(부분 익절 포함)의 순실현이며 `closed_at` 에 귀속된다. 동결 거절 20 종은 원장에 남지 않는다. a130 backfill 은 **체결로 닫힌 편입 포지션**만 되메운다 — 엔진 진입 포지션의 거절(`cost_estimate_failed` · `insert_failed` · `round_trip_unpriceable` …)은 복구 경로가 없고, 조정(adjustment)으로 닫힌 포지션은 대상 밖이다 | trade_outcomes.go · trade_outcome_refusal.go · 독립 대조 P1-5 · P2-6 |
| F10 | 일 경계: 일일 손실 예약의 시장 현지일 소멸(reservation_release.go:379-455)은 **생산에서 실행되는 일이 없다**(정정). 살아 있는 선례는 콘솔 `marketDayStart`(overview.go:917)와 `journal.ReadOnly.AccountTradeTrips`(account_views.go:592)다 | 독립 대조 P2-7 |
| F11 | 공식 Open API 에 실현 손익 조회가 없고 엔진은 official-only 다 | `internal/official` endpoint 열거 |
| F12 | 원장은 단일 연결(`SetMaxOpenConns(1)`)이다. 행을 연 채 네트워크를 타면 모든 원장 쓰기가 막힌다. `PrecheckQFinal*` 은 「원장 변경 · 브로커 호출 없음」 계약이다 | journal 열기 · riskguardian_qfinal.go:79-81 |
| F13 | 조정으로 닫힌 포지션: 조정이 수량을 0 으로 만들면 같은 트랜잭션이 `positions.state = CLOSED` · `closed_at = now` 로 바꾸고 `position_adjustments` 에 `new_quantity` 0 행이 남는다 | position_adjustments.go:332-352 · `convergedState` |

## 사용자 결정 (2026-10-14 최종) — 설계 입력

| Q | 결정 |
|---|---|
| Q1 | 원천 = 원장. riskcalc `RealizedPnL.AsOf` 주석을 고친다 |
| Q2 | 열린 포지션의 부분 청산은 제외(명시 한계, 후속) |
| Q3 | 성과 행 없는 오늘의 청산 → 그 시장일 fail-closed. 엔진 진입 포지션의 거절은 복구 경로가 없으므로 **일 경계에서만** 풀린다(F9) |
| Q4 | 하루 = KST 08:00 에 넘어가는 하루 ≡ 각 시장 자기 거래일 D, 시장을 가로질러 합산. a136 D1 과 같은 정의이고 a136 이 일 키 함수를 한 곳에 둔다 — 그것을 부르고 복제하지 않는다 |
| Q5 | 환율 = 같은 파도에서 US 쪽이 이미 봉인한 USD/KRW 권한 재사용(새 브로커 호출 0). 스펙 준수를 확인하고, 정말 못 지키면 멈추고 보고 → **D5** |
| Q6 | 전략 경로의 일일 손실 거절 = 계좌 단위 진입 차단, ENTRY_BLOCKED 로 강화. 원인 보존(F8) |
| Q7 | 강화는 사람이 풀 때까지(현 스펙 유지) |
| Q8 | 미실현 제외 |
| Q12 | 그날 닫힌 포지션의 **손실만** 센다. 이익은 상쇄하지 않는다 — 하루 안에서 측정 합계는 줄지 않는다 |
| Q13 | 사람이 앱에서 팔아 **조정으로 닫힌** 관리 포지션은 손실로 세지 않고 fail-closed 도 하지 않는다(명시 한계) — 「수동 주문 손실은 세지 않음」과 일관 |

## D1. 측정 모집단과 산식

일 키 `D` 는 a136 의 공유 함수가 준다(KST 08:00 롤오버). 측정기는 `D` 를 입력으로 받고 직접 계산하지 않는다.

1. 계좌의 `positions` 중 `state = CLOSED` 이고 시장이 KR · US 이며 `clock.Market.TradingDay(closed_at) == D`(그 포지션 **자기 시장**의 현지일)인 것.
2. 제외: 관리 대상이 아닌 포지션(`ExitEligible` 거짓 — 진입 결정도 편입도 없는 외부 · 수동 보유) · 조정으로 닫힌 포지션(F13, Q13).
3. 남은 각 포지션: 성과 행이 있으면 `realized_pnl_after_costs`, 없으면 **측정 불가**(Q3 → 진입 거절).
4. Q12: 금액이 음수인 것만 남긴다. 이익 · 0 은 버린다.
5. 각 손실을 계좌 기준 통화로 환산(D5). KR → KRW, US → USD 는 포지션 시장에서 유도(성과 행에 통화 칸이 없다).
6. `riskcalc.DailyLoss{TradingDay: D, TargetCurrency: base, Realized: [{TradingDay: D, Amount: 손실(음수, 기준 통화), Currency: base, CostsDeducted: true, AsOf: 읽은 순간}]}`.
   공식은 riskcalc 그대로 — 음수만 넣으므로 `max(0, −순실현)` = 손실 절댓값 합. `AsOf` 는 체인 `Now` 이하(F4).

성과의 `closed_at` 과 포지션의 `closed_at` 은 같은 체결 트랜잭션에서 쓰인다. 모집단 판정은 포지션 쪽 하나로 한다(성과 행이 없는 포지션도 같은 기준으로 셈).

## D2. 측정의 자리 — Guardian, 체인 직전, 한 번

- `PrecheckQFinalEntry`(살아 있는 입구)와 `IssueEntry`(잠든 입구 — 나중에 깨어나도 같은 바닥을 보도록) 모두 체인 직전에
  `Account.DailyRealizedLoss = max(호출자 값, 측정값)` 으로 바꾼다. 호출자 값을 낮추지 않는다.
- 측정 실패: 측정 오류를 들고 값을 쓸 수 없는 Money 로 둔 채 체인을 돈다. 체인이 `daily_loss` 단에서 `INPUT_UNAVAILABLE` 로 멈추면 그 `Detail` 을
  측정 오류 문구로 바꾼다(계좌 값 · 원문 오류 없음 — a131/a133 경계). 앞 단(kill switch · 운영 모드 · latch …)이 먼저 거절하면 그 사유가 그대로다.
- **ctx 문제와 측정 토큰.** `PrecheckQFinalEntry` 는 ctx 를 받지 않고 원장 · 브로커를 건드리지 않는 계약이다(F12). 그래서 원장 읽기는 precheck
  **앞에서** Guardian 의 새 메서드 `MeasureDailyLoss(ctx, 일 키)` 가 하고, 결과를 **불투명 토큰**(내보내지 않는 칸만 — `QFinalEntryPrecheck` 와 같은 패턴,
  다른 패키지가 만들 수 없음)으로 돌려준다. 토큰은 통화별 손실 합(아직 환산 전) · 일 키 · 측정 순간 · 발급한 Guardian 신원을 봉인한다.
  첫 다리 조립(strategy_account_first_leg_authority.go)이 토큰을 `QFinalEntryIssuance` 에 실어 넘기고, precheck 가 ① 토큰이 이 Guardian 의 것이고
  ② 측정 순간이 Guardian 시계 기준 계좌 스냅샷 신선 한계(10초) 안이며 미래가 아니고 ③ 일 키가 지금의 일 키와 같은지 확인한 뒤 요청의 봉인 FX 로 환산한다.
  토큰이 없거나(영값) 하나라도 어긋나면 측정 실패 → `daily_loss` 단 `INPUT_UNAVAILABLE`.
- 측정은 precheck 에서 **한 번**이다. `IssuePrecheckedQFinal*` 은 다시 재지 않는다 — precheck 와 발급 사이(결정 TTL 안)에 닫힌 손실은 그 한 건이
  못 보고 다음 진입이 본다(명시 한계).
- 원장 읽기는 계좌 색인으로 한정한 한 번의 질의이고, 행을 연 채 환산 · 네트워크를 하지 않는다(F12). 환산은 이미 봉인된 권한의 산술뿐이라
  브로커 호출이 없다 — `PrecheckQFinal*` 의 「브로커 호출 없음」 계약 유지.
- `IssueReduction` · `evaluateReduction` · 청산 · 손절 경로는 측정기를 부르지 않는다(시험으로 고정). `execgw.Issuer.IssueEntry`(체인 없는 문)는
  생산 호출자 0 을 시험으로 고정한다.
- 게이트 OFF: 생산 Guardian 이 만들어지지 않으므로 측정도 없다(토글 OFF = upstream 동작).

## D3. 전략 경로의 분류와 강화 (Q6)

- `PrecheckQFinalEntry` 는 원장을 바꾸지 않는 계약이라 강화를 안에서 하지 않는다. 첫 다리 승인 함수(strategy_first_leg_admission.go)가 precheck 거절에서
  `*IssueRefusal` 의 reason 이 `DAILY_LOSS_LIMIT_REACHED` 임을 **타입으로** 보면(문구로 가르지 않음) Guardian 의 새 강화 입구(ctx, 원장 쓰기)를
  부르고, 계좌 단위 진입 차단 코드로 돌려준다. 원인은 `cause` 로 싣는다(F8). 이 분류는 레인 결함 · 범위 결함이 아니다.
- 강화 뒤에는 운영 모드가 ENTRY_BLOCKED 라 뒤 진입은 앞 단(`operating_mode`)에서 거절된다. 그 주기의 나머지 범위를 계속 돌지, 멈출지는 FLM 뒤
  확정한다(tasks 0.7 — `deliverEachStrategyHandoff` 의 「범위 거절만 건너뜀」 규칙과의 관계).
- Q7: ENTRY_BLOCKED 는 사람 해제까지 유지된다(일 경계에서 풀리지 않음).

## D4. 원장이 세지 못하는 손실 (명시 한계)

| # | 빠지는 손실 | 결정 |
|---|---|---|
| U1 | 외부 수동 거래 · 관리 대상이 아닌 포지션 | 세지 않음(F11 — 공식 원천 없음) |
| U2 | 열린 포지션의 부분 청산 | Q2 제외. 단 그 포지션이 그날 완전히 닫히면 성과 행이 수명 전체를 담아 그날에 들어간다(F9) |
| U3 | 조정으로 닫힌 관리 포지션(사람의 앱 매도) | Q13 세지 않음, fail-closed 아님 |
| U4 | 브로커는 체결했으나 원장 미반영 | 기존 「체결 감지 5주기 실패 → ENTRY_BLOCKED」가 맡음 |
| U5 | 미실현 | Q8 제외 |
| U6 | 이익으로 상쇄된 손실 | Q12 — 상쇄하지 않으므로 빠지지 않음(보수) |
| U7 | precheck 뒤 · 발급 전에 닫힌 손실 | 그 한 건은 못 봄, 다음 진입이 봄(D2) |

## D5. 환율 — Q5 의 스펙 준수 판정

**판정: 지금 문장 그대로는 준수하지 못한다. 좁은 스펙 수정(MODIFIED) 하나로 준수할 수 있다.** 그래서 멈추지 않고 수정안을 델타에 넣었고,
그 수정은 사람 승인 대상이다(review.md 「남은 사람 승인」).

- 근거 문장: risk-management 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」 — 각 exposure-raising 요청은 **자기** request-scoped frozen FX
  authority 를 쓰고, 「같은 authority가 Guardian sizing, aggregate와 … 끝까지 사용되어야 하며 중간 재조회나 **다른 환율 사용은 금지**」(MUST NOT).
- US 진입: 오늘의 US 손실(USD)은 그 요청 **자기** 권한(USD→기준)으로 환산된다 — 문장 그대로 준수. KR 손실은 기준 통화(KRW 기준이면 환산 없음).
- KR 진입: 요청 권한은 identity(KRW/KRW)뿐이다. 오늘의 US 손실을 같은 파도의 US 권한으로 환산하면 KR 요청의 aggregate 가 「다른 환율」을 쓴다 → 문장 위반.
- 수정안: 「일일 손실 측정에서 요청 시장과 다른 시장의 실현 손실은 **같은 파도에서 공식 원천이 봉인한 그 시장의 권한**으로만 환산하며, 그 권한이 없으면
  fail-closed 한다. 재조회 · 호출자 환율 · 다른 파도의 권한은 금지」를 그 요구에 더한다. 재조회 금지 · 공식 원천 · 봉인이라는 요구의 목적은 그대로다.
- 권한 부재: US FX 권한은 US 후보 스냅샷이 Ready 일 때만 봉인된다(strategy_fx_authority.go:205). 없으면 그 KR 진입은 거절된다.
- 실제 빈도(관찰, 측정 아님): 하루 키가 KST 08:00 롤오버이고 US 손실은 **US 현지일** D 로 묶이므로, KR 정규장(09:00~15:30 KST, 미 동부 전날 20:00~02:30) 동안
  US 현지일 D 의 청산은 미국 확장 시간대 청산뿐이다. KR 진입이 USD 환산을 요구하는 경우는 드물다.
- 공유 표면: 권한을 KR precheck 까지 운반하려면 `QFinalEntryIssuance`(riskguardian_qfinal.go — a136 이 같은 파일의 `IssuePrecheckedQFinalEntry` 를 편집)와
  첫 다리 조립(strategy_account_first_leg_authority.go)에 칸 하나가 필요하다 → a136 커밋 뒤.

## Fail-closed 가 거부하는 정상 입력 (결정 반영, 진입만)

1. 원장 읽기 오류(busy · 잠금 · 닫힌 핸들) — 그 시도.
2. 그날(D) 닫힌 관리 포지션 중 성과 행이 없는 것 — **그 일 키의 나머지 진입 전부**. 편입 포지션은 a130 backfill 로 풀 수 있고, 엔진 진입 포지션
   (`cost_estimate_failed` · `insert_failed` · `round_trip_unpriceable` · `sell_unpriced` · `sell_quantity_mismatch` …)은 **다음 일 경계까지** 풀리지 않는다.
3. KR 진입 시 그날의 US 손실이 있는데 같은 파도의 US FX 권한이 없음(D5).
4. `closed_at` · 성과 금액 해석 불가 — 그 일 키의 진입 전부(원장 값 결함, 사람 조사).
5. 시장 시간대 로드 실패 — 그 시장의 진입.

거부하지 않아 통과하는 것: U1 · U2 · U3 · U4 · U5 · U7. 매도 · 손절 · 비상 청산 · 대사 · 체결 감지는 측정기를 부르지 않으므로 어느 것도 거부되지 않는다.

## 구현 순서

1. (지금, 새 파일만) 측정기 순수 함수 패키지 · 원장 읽기 새 파일의 RED 시험(빌드 태그 `a137red` 아래 — 공유 체크아웃의 `go test ./...` 를 깨지 않음) ·
   엔진 쪽 편집 대상 함수의 FLM/BTM(a136 소유 파일 밖) · proposal-freeze 리뷰.
2. a136 공유 파일 커밋 확인(`git log -1 -- <file>` · `git diff`) → base = 이 change 첫 Go 커밋의 부모 → riskguardian*.go FLM 재추출 → Pre-Edit → GREEN.
3. 변이 원장 → 적대적 독립 리뷰 → Manager 게이트.
