# a138 evidence reconciliation

| 항목 | CodeGraph | HEAD 읽기 · AST | 결론 |
|---|---|---|---|
| 예약 tx 진입점 | `reserveRows` 호출자 5(생산) | AST `journal--reserverows` B3 `checkLimits` → B4 `checkDailyBuyNotional` → B5 insert | 일치 — 훅 한 자리로 진입 발급 문 셋 모두 덮음 |
| 매도 경로 | `IssueReduction` 은 `reserveRows` 호출자 목록에 없음 | AST `execgw--riskguardian.issuereduction` 호출 = `RecordDecision` · `evaluateChain` 등, 예약 API 없음 | 일치 — 예약 tx 판정은 매도에 닿지 않음 |
| 발급 문 수 | `dailyBuyLimit` 호출자 3 | 세 문 모두 OPEN_EXPOSURE 한 종류 예약(a136 근거) | 일치 — 같은 자리에 상한 한 칸씩 |
| 정책 유도 | `riskPolicyFromAutomationGate` 생산 호출자 1 | AST `engine--riskpolicyfromautomationgate` 분기 0 · 반환 2(39 는 클로저) | 일치 — 직선 함수에 대입 한 줄 |
| 인터록 | `checkGuardianDailyBuyLimit` 호출자 `verifyGate` | AST `engine--verifygate` B1~B10, 7b(하루 매수)가 B8(534) | 일치 — B8 다음에 한 단 |
| 정책 검증 | — | AST `risk--policy.validate` B1~B11 | 정수 상한 검사는 새 분기(편집 시 FLM) |

불일치 0(아래 정정 1 건 추가).

정정(2026-10-15, a139 조사 중 발견): baseline 표의 `risk.DefaultPolicy` 「생산 2(guardian_wiring · measure)」는 HEAD grep 근거다 — CodeGraph callers(raw/callers-DefaultPolicy.txt)는
`riskPolicyFromAutomationGate`(guardian_wiring.go:47)의 호출을 빠뜨리고 `gridPolicy` 하나만 낸다. 불일치 1, HEAD 가 권위.
