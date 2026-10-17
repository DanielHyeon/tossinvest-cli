# a139 evidence reconciliation

| 항목 | CodeGraph | HEAD 읽기 · AST | 결론 |
|---|---|---|---|
| `DefaultPolicy` 생산 호출자 | `gridPolicy` 하나 | grep: `guardian_wiring.go:47`(`riskPolicyFromAutomationGate` 안) · `counterfactual.go:354` | **불일치** — CodeGraph 가 `riskPolicyFromAutomationGate` 의 호출을 빠뜨림. HEAD 가 권위: 생산 호출자 2 |
| `Policy.RiskBudget` 읽는 자리 | sizing 호출자 사슬(raw/callers-*) | grep 전수 + AST(design D0) | 일치 — 사이징 셋 · 검증 · 유도 둘 · 격자 |
| digest · 인터록 · 콘솔 | — | AST `execgw--riskguardian.limitsdigest` 입력에 없음, `checkGuardianLimits` 다섯 칸, console grep 0 | RiskBudget 을 읽지 않음 → 바인딩 무효화 없음 |
| 매도 경로 | — | AST `risk--evaluate` B2(135) · `risk--preflight` B6(180) | 매도는 사이징 단 전에 분기 |

불일치 1(CodeGraph 누락) — HEAD 로 해소. a138 baseline 의 「생산 2(guardian_wiring · measure)」 문장도 같은 grep 근거이며 CodeGraph 단독 근거가 아님(a138 reconciliation 에 같은 정정 기록).
