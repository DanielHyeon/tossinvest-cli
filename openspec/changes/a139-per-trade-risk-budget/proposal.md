# a139 — 한 거래의 위험 예산은 일일 손실 한도의 1/3 (per-trade-risk-budget)

> **상태(2026-10-15):** proposal 단계, 문서만. 사용자 결정 「RiskBudget = MaxDailyLoss / 3」 수신(Manager 전달). production 파일은 a138 의 Go · spec 이
> 전부 커밋된 뒤에 만진다(Manager 지시 — 공유: `riskPolicyFromAutomationGate` · `risk.DefaultPolicy` · risk-management 「Automation gate limits …」).

## Why

지금 `policy.RiskBudget = dailyLoss`(guardian_wiring.go:59)라 손절에 걸린 거래 **하나**가 그날의 손실 한도를 다 쓸 수 있다. risk-management
「Automation gate limits are the production Guardian policy source」가 그것을 규범으로 적고 있다(「The per-trade risk budget SHALL equal the configured
maximum daily-loss amount」, spec.md:165). 사용자 결정(2026-10-15): 건당 위험 예산 = 일일 손실 한도 / 3. 사이징을 **줄이는** 방향만이다(§0.9).

## What Changes

1. **유도 한 곳** — 새 함수 `risk.PerTradeRiskBudget(dailyLoss riskcalc.Money) (riskcalc.Money, error)`: `floor(dailyLoss / 3)` 을 그 통화의 최소 단위(KRW 1 ·
   USD 0.01 — 저장소의 `minor-v1` 관례, continuationlane `currencyMinorScale` 과 같은 표)로 내림. 최소 단위를 모르는 통화 · 0 으로 내려가는 값 · 해석 불가는 오류.
2. **생산 정책** — `riskPolicyFromAutomationGate` 가 그 함수를 부름. 오류면 예산을 비워 `Policy.Validate` 가 거부 → 게이트 ON 기동이 `ErrLimitsRequired` 로 멈춤
   (a136 이 무효 값을 검증에 넘긴 것과 같은 길 — 이 함수는 오류를 돌려줄 자리가 없다).
3. **기본 정책** — `risk.DefaultPolicy().RiskBudget` 100,000 → 33,333 KRW(같은 규칙 — 주석의 「RiskBudget = MaxDailyLoss」 유도 문장을 고침). 생산에서는 덮어쓰이므로 행동
   영향은 시험뿐(D3).
4. **감사** — 기동 audit 끝에 `engine.automation_gate.per_trade_risk_budget`(유도 값) 한 줄, 인터록 acceptance 문구에 `per_trade_risk_budget=`.
5. **spec** — `risk-management` 「Automation gate limits …」 MODIFIED(1/3 · 내림 · provenance · 시나리오 2). a138 델타 문장 위에 씀.

## What Does NOT Change

- 사이징 공식 `floor(위험예산 / (entry − stop))`, 다른 사이징 상한(주문 수량 · notional), 체인 순서 · reason-code.
- `LimitsDigest` · 결정 `limits_json` 5필드 · 인터록 5필드 대조 — RiskBudget 은 그 어디에도 들어 있지 않다(AST `execgw--riskguardian.limitsdigest` 입력 =
  limitsJSON + 하루 매수 상한, D2). 따라서 생산 전략 바인딩 무효화 없음.
- `productionGuardianPolicyVersion`(D4 — 올리지 않음).
- 격자 측정(`measure.gridPolicy`)의 예산 = headroom(기하만 거부하는 측정용 — 생산 사이징이 아님).
- 매도 · 손절 · 비상 청산 · 대사 · 체결 감지. 원장 스키마 · 설정 키(새 키 없음 — 1/3 은 설정이 아니라 정책).
- 전략 레인의 캠페인 예산(`RiskBudgetMinor` — 서명 제안서에서 오는 별개 값).

## Impact

- Affected specs: `risk-management`(MODIFIED 1).
- Affected code(예정): `internal/risk/input.go`(DefaultPolicy) · 새 `internal/risk/risk_budget.go`, `internal/app/engine/guardian_wiring.go` · `interlock.go`
  (acceptanceDetail · recordGateSettings).
- 행동: 생산 진입의 위험 기반 수량 상한이 1/3 로 줄어든다 — 체인 `checkOrderSize`(의도 수량이 예산 상한을 넘으면 INVALID_ORDER_SIZE) · `StrategyEntryQuantity` ·
  `AccountBaseStrategyEntryQuantity`(q_final Guardian 상한). 예산이 한 단위 위험도 못 덮는 넓은 손절은 수량 0 → 거부가 늘어난다(보수 방향).
- 위험 등급: **High-risk**(사이징). Pre-Edit · FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰.
- 배포: 엔진 이미지 교체. 설정 · 서명 매니페스트 변경 불요.
- **archive 순서**: a138 이 먼저 archive 되어야 이 델타(a138 문장 포함)가 충돌 없이 합쳐진다.
