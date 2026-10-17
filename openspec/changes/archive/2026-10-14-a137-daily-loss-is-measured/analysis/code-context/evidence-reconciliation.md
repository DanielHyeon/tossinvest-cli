# a137 evidence reconciliation

기준 HEAD `7b073f55` · codegraph 1.6.0 · tools/logic-map AST(`analysis/ast-evidence/`) · HEAD 직접 읽기.

| # | 주장 | CodeGraph | AST / HEAD | 판정 |
|---|---|---|---|---|
| R1 | `riskcalc.DailyLoss` 생산 호출자 0 | callers 6 전부 시험 | `grep -rn "DailyLoss(" internal cmd` 비시험 = 정의뿐 | 일치 |
| R2 | 체인 단은 값이 없으면 막고 0 이면 통과 | callers `entryChain` | `risk--checkdailyloss.ast.json` B4~B9 · `risk--magnitudein.ast.json` B1~B3 | 일치 |
| R3 | 매도 의도는 daily_loss 단에 오지 않음 | impact 5(Evaluate 포함) | `risk--evaluate.ast.json` B2(135) `SideSell` → evaluateReduction | 일치 |
| R4 | 서명 계좌 매니페스트가 생산 전략 경로의 값 | `LoadProductionAuthority` 호출자 = 시험 2 | strategy_account_first_leg_authority.go:116 함수 값 참조 · :388 `Account: accountAuthority.AccountState()` | **불일치(도구 한계)** — 함수 값 참조를 호출로 세지 않음. HEAD 로 해소 |
| R5 | tracer 는 리터럴 0 이고 생산 생성자 0 | `accountState` 호출자 1 | `engine--tracer.accountstate.ast.json`(분기 1 = 통화) · `grep -rnF "NewTracer(" cmd internal` 비시험 = 정의뿐 | 일치 |
| R6 | 자동 강화는 `IssueEntry` 에만 | `escalateFor` 호출자 1 | `execgw--riskguardian.precheckqfinalentry.ast.json` B16(158) 거절 = `chainRefusal` 만 · `grep escalateFor internal/execgw` 비시험 호출 1(riskguardian.go:436) | 일치 |
| R7 | 매니페스트 생산자가 저장소에 없음 | — | `grep -rln "daily_realized_loss_base\|strategy-account-snapshot"` → production.go · production_test.go · a112 review.md · 빌드 산출물 둘 | 일치(저장소 범위만 — 운영 config 디렉터리는 읽지 않음) |
| R8 | `RiskBudget = MaxDailyLoss` 는 의도된 건당 예산 | — | `engine--riskpolicyfromautomationgate.ast.json`(분기 0) · risk-management 「Automation gate limits …」 | 일치 — Manager 갭 문장의 둘째 절은 결손이 아니라 스펙 |
| R9 | 전략 경로 한도 거절의 하류가 레인 잠금 | — | strategy_first_leg_admission.go:93(`AUTHORITY_MISMATCH`, 범위 거절 아님) → strategy_market_handoff_delivery.go:69(범위 거절만 건너뜀) → strategy_entry_supervisor.go:966 `latchMarket` | **미고정** — HEAD 읽기만, refreshOnly 갈래(:940-960) 와의 경합 미확인. 구현 전 FLM(tasks 0.7) |

불일치 R4 는 도구 한계이고 HEAD 로 해소했다. R9 는 편집 차단 사유(구현 전 AST 로 고정할 것).
