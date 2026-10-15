# a139 CodeGraph baseline

- 도구: codegraph 1.6.0, 질의 시점 HEAD `050d092f`(코드 = a1d66f13 — 그 사이 커밋은 a138 문서뿐). 원문은 `raw/`.

| 심볼 | 생산 호출자 | a139 영향 |
|---|---|---|
| `risk.accountBaseRiskQuantity` | `AccountBaseStrategyEntryQuantity` · `checkOrderSize` | 예산 ÷ 손절 폭 — 공식 무편집, 들어오는 예산만 1/3 |
| `risk.checkOrderSize` | `entryChain`(chain.go:96) | 체인 「주문 크기 한도」 — 무편집 |
| `risk.StrategyEntryQuantity` | `RiskGuardian.IssueStrategyEntry` · `PrecheckQFinalEntry` | 무편집 |
| `risk.AccountBaseStrategyEntryQuantity` | `PrecheckQFinalEntry` | 무편집 |
| `engine.riskPolicyFromAutomationGate` | `newProductionRiskGuardian` + 시험 | **편집** — 유도 함수 호출 |
| `risk.DefaultPolicy` | `measure.gridPolicy` (CodeGraph) · HEAD grep 은 `guardian_wiring.go:47` 도 찾음 | **편집** — 33,333 |
| `engine.recordGateSettings` · `acceptanceDetail` | `NewContext` · `runInterlock` | **편집** — audit 한 줄 · 문구 한 칸 |

`codegraph affected <3 files> --filter '*_test.go'` → 1133 줄. `grep-RiskBudget-nontest.txt` 는 `RiskBudget` 철자 전수(레인 `RiskBudgetMinor` 포함 96 줄) — design D0 이 그중
`risk.Policy.RiskBudget` 을 읽는 자리만 추림.
