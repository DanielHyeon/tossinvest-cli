# a139 design — 건당 위험 예산 = 일일 손실 / 3

근거 표기: 「AST:」는 `analysis/ast-evidence/<번들>.ast.json`(`tools/logic-map`, HEAD ddd2f2b3 — a138 착지 · archive 뒤 재추출, 편집 전).
CodeGraph 근거는 `analysis/code-context/`.

## 사용자 결정 (2026-10-15 — Manager 전달 원문)

> a139-per-trade-risk-budget (user decision 2026-10-15: RiskBudget = MaxDailyLoss / 3)
> - Today `policy.RiskBudget = dailyLoss` … so one trade may risk the whole daily-loss budget. Change to one third (conservative sizing direction only —
>   invariant 6). Read the spec requirement that defines RiskBudget … and MODIFY it with provenance (정책 수치의 provenance). Decimal rounding must be
>   conservative (down). Check every consumer of RiskBudget with CodeGraph + AST (sizing, envelope, digest/LimitsDigest, interlock, console display) and
>   pin behaviour with tests.

## D0. `Policy.RiskBudget` 의 소비자 전수

(a138 착지 뒤 재확인 — a138 은 `Policy` 에 `MaxOpenPositions` 를 더했을 뿐 `RiskBudget` 소비자를 바꾸지 않음. 전수 grep 재실행 결과 같음.)

`rtk proxy grep -rn RiskBudget --include=*.go`(시험 제외)에서 `risk.Policy.RiskBudget` 을 읽는 생산 자리만 추림(레인의 `RiskBudgetMinor` · `RiskBudgetDigest` 는
서명 제안서의 캠페인 예산 — 다른 값, 이 change 밖).

| 소비자 | 무엇을 하나 | 근거 |
|---|---|---|
| `risk.accountBaseRiskQuantity` | `floor(budget / ((entry − stop) × FX 승수))` — 예산이 음수 · 해석 불가면 오류, 폭 ≤ 0 이면 0 | AST `risk--accountbaseriskquantity` B2(220) · B5(232) |
| `risk.checkOrderSize`(체인 「주문 크기 한도」) | 유도 · 해석 오류면 INPUT_UNAVAILABLE(B4 334 · B5 338), 값이 0 이면 INVALID_ORDER_SIZE(「예산이 한 단위 위험도 못 덮음」, B6 341), 의도 수량 > 그 값이면 INVALID_ORDER_SIZE(B7 346) | AST `risk--checkordersize`, chain.go:333-351 |
| `risk.StrategyEntryQuantity` | `min(위험 상한, 주문 수량, notional)` — 같은 통화 경로 | AST `risk--strategyentryquantity`(호출 `RiskBasedQuantity` · `policy.Validate`) |
| `risk.AccountBaseStrategyEntryQuantity` | 같은 min, KR/US 짝 FX 경로 | AST `risk--accountbasestrategyentryquantity`(호출 `accountBaseRiskQuantity`) |
| `execgw.RiskGuardian.IssueStrategyEntry` | `StrategyEntryQuantity` 로 수량 | riskguardian.go:357 |
| `execgw.RiskGuardian.PrecheckQFinalEntry` | Guardian 상한(`q_existing_guardian`) = 위 둘 중 하나 | AST `execgw--riskguardian.precheckqfinalentry` B11(142) · B12(144), qfinal.go:141-149 |
| `risk.Policy.Validate` | 예산이 양수 · 통화 일치여야 함 | AST `risk--policy.validate` B2 · B3(267 · 276 — 돈 칸 목록 순회) |
| `risk.DefaultPolicy` | 100,000 KRW(= 일일 손실) | AST `risk--defaultpolicy`(분기 0) · input.go:228-254 |
| `engine.riskPolicyFromAutomationGate` | `= dailyLoss` | AST `engine--riskpolicyfromautomationgate`(분기 0), guardian_wiring.go:59 |
| `measure.gridPolicy` | `= headroom`(측정 격자 — 생산 사이징 아님) | counterfactual.go:367 |

읽지 **않는** 곳(매니저가 지목한 표면): `LimitsDigest`(AST `execgw--riskguardian.limitsdigest` — 입력은 `limitsJSON` + 하루 매수 상한뿐), 결정 `limits_json`
5필드(`execgw.ExposureLimitsFor` — 다섯 칸), 인터록 `checkGuardianLimits`(다섯 칸), 콘솔(`internal/console` · `cmd/tossctl` grep 0). envelope(q_final
`riskbucket` 정책)은 서명 위험 번들의 정책이고 `Policy.RiskBudget` 을 읽지 않는다(grep 0).

## D1. 내림의 단위 — 통화의 최소 단위

1/3 은 유한 소수가 아니므로 반드시 자른다. 안:

| 안 | 100,000 KRW | 50 USD | 50.75 USD | 비고 |
|---|---|---|---|---|
| 정수 단위로 내림 | 33,333 | 16 | 16 | 표 불필요, 더 보수적. USD 에서 최대 0.99 를 버림 |
| **통화 최소 단위로 내림(채택)** | 33,333 | 16.66 | 16.91 | spec 「KR과 US …」가 risk budget 을 「base minor unit 으로 비교」하라고 적음. 저장소 관례 `minor-v1`(KRW 0 · USD 2자리) |
| 입력 소수 자릿수로 내림 | 33,333 | 16(「50」) | 16.91 | 철자 의존(50 과 50.00 이 다른 답) — 기각 |

채택: 최소 단위 내림. 표는 `risk` 안의 작은 사본(KRW 0 · USD 2)이며 continuationlane `currencyMinorScale` 과 같은 값이다(그 함수는 비공개 · 다른 패키지이고
그 레인의 호가 단위 표라 옮기지 않음 — 두 표는 서로를 대조하지 않는다. `risk` 쪽 값은 시험이 고정한다). 표에 없는 통화는 오류 → 기동 거부(그 통화는 지금도 KR/US FX authority 가 없어 자동 진입 권위가 없다 —
console `currencyConsequence`). 1/3 이 0 으로 내려가는 값(KRW < 3, USD < 0.03)도 오류 → 기동 거부(`unsupported` 극단 입력, 명시).

## D2. 바인딩 · digest · 인터록

RiskBudget 은 digest · 결정 스냅숏 · 인터록 대조 어디에도 없다(D0). 그래서 이 change 는 서명 매니페스트 · 전략 바인딩을 무효화하지 않는다. 대신 유도 값이
**감사에** 남게 한다: `recordGateSettings` 끝에 `engine.automation_gate.per_trade_risk_budget` 한 줄(a136 · a138 과 같은 덧붙임 규칙 — 기존 순서 무변경),
`acceptanceDetail` 에 `per_trade_risk_budget=`. 인터록에 새 대조 단은 두지 않는다 — 생산 Guardian 은 정책을 이 유도에서만 얻고(`newProductionRiskGuardian`), 주입된 다른
Guardian 의 예산은 지금도 대조하지 않는다(잔재로 기록, 이 change 가 넓히지 않음).

## D3. 기본 정책도 같은 규칙

`DefaultPolicy` 의 예산은 생산에서 `riskPolicyFromAutomationGate` 가 덮어쓰므로 행동 영향은 시험뿐이지만, 주석이 「RiskBudget = MaxDailyLoss」를 규칙으로 적고 있어
두 곳이 다른 규칙을 말하게 둘 수 없다(문서 · 코드 · 시험 일치). 값은 33,333 KRW 로 적고, 시험이 `DefaultPolicy().RiskBudget == PerTradeRiskBudget(DefaultPolicy().MaxDailyLoss)` 를
못 박는다(사본이 갈라지지 않게). DefaultPolicy 에 기대는 시험의 기대 수량이 바뀌는지는 RED 단계에서 세어 기록한다.

## D4. 정책 버전은 올리지 않음

`productionGuardianPolicyVersion`(risk-policy-v1)은 결정 preimage 와 q_final 정책 버전에 들어간다. 생산 첫 다리 경로는 기대 버전을 Guardian 자신에게서 읽고
(strategy_account_first_leg_authority.go:393), 휴면 어댑터만 매니페스트 값과 비교한다(strategydispatch/adapters.go:33). 올리면 문자열만 바뀌고 안전 판정은 그대로이며,
휴면 경로의 매니페스트를 깨는 비용이 생긴다. 사이징 규칙의 변경 기록은 이미지 · 커밋 · 기동 audit 한 줄(D2)이 맡는다. — freeze 리뷰가 다르게 보면 바꾼다.

## D5. 거절되는 정상 입력

- 손절 폭이 넓어 1/3 예산이 한 단위 위험도 못 덮는 진입 — 체인 INVALID_ORDER_SIZE / q_final `q_existing_guardian` 0 → 거부(이전에는 통과하던 것 일부).
- 전략이 제안한 수량이 1/3 예산 상한보다 큰 진입 — 체인 경로는 INVALID_ORDER_SIZE 거부, Guardian-owned 사이징 경로는 수량이 줄어듦.
- 일일 손실 < 3 KRW · < 0.03 USD 인 게이트 — 기동 거부.

## D6. 매도는 무관

`RiskBudget` 은 진입 사이징에서만 읽힌다(D0). 위험 축소 발급(`IssueReduction`)은 체인을 부르지만 매도는 `Evaluate` 의 B2(135 — `Side == SideSell` →
`evaluateReduction`)로 가 진입 체인(`checkOrderSize` 포함)을 지나지 않고(AST `risk--evaluate`), 그 앞의 `preflight` 도 매도는 정책 검사 전에 허용한다(chain.go
preflight — AST `risk--preflight` B6(180) `Side != SideBuy` → allow). 예산 유도가 실패하는 설정은 Guardian 이 만들어지기 전에 기동이 멈추므로(`newProductionRiskGuardian` 의 `Validate`)
가동 중인 엔진에 무효 예산이 존재하지 않는다.
