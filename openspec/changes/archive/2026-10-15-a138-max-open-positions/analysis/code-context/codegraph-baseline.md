# a138 CodeGraph baseline

- 도구: codegraph 1.6.0, 질의 시점 HEAD `a1d66f13`(워킹트리 변경은 a128 프로브 원장뿐 — 코드 무변경). 원문은 `raw/`.

| 심볼 | 생산 호출자 (callers) | a138 영향 |
|---|---|---|
| `journal.reserveRows` | `RecordDecisionAndReserve`(issuance.go:106) · `RecordQFinalCampaignFirstLeg`(strategy_first_leg_atomic.go:101) · `RecordStrategyDecisionAndReserve`(strategy_lineage.go:315) · `RecordQFinalDecisionAndReserve`(risk_bucket_issuance.go:70) · `Reserve`(reservations.go:251) + 시험 1 | **모든 예약 tx 가 이 한 함수를 지난다** → 판정 훅 한 자리(a136 5b 다음) |
| `journal.checkDailyBuyNotional` | `reserveRows` 하나 | 선례 — 같은 모양의 새 함수 `checkOpenPositions` 를 옆에 둠(무편집) |
| `execgw.RiskGuardian.dailyBuyLimit` | `IssueEntry` · `IssuePrecheckedQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` | 발급 문 셋 = 상한을 싣는 자리 셋 |
| `RiskGuardian.LimitsDigest` | `IssueStrategyEntry` · `PrecheckQFinalEntry` · strategydispatch `IssueAndPlan` + 시험 | digest 입력에 상한 추가 → 바인딩 결속 |
| `riskPolicyFromAutomationGate` | `newProductionRiskGuardian` + 시험 3 | 정책 유도 한 곳(a139 와 공유) |
| `checkGuardianDailyBuyLimit` | `verifyGate` | 선례 — 새 `checkGuardianOpenPositionLimit` 를 옆에(verifyGate 에 한 단 추가) |
| `newAutomationStatus` · `acceptanceDetail` · `recordGateSettings` | engine.go:437(NewContext) 등 | 해석 값 운반 · 감사 문구 · 기동 audit 한 줄 |
| `config.Service.SaveEngineGateLimits` · `gateMembersOf` | cmd/tossctl limitsettings · 시험 | 콘솔 저장 seam — 새 키 멤버 |
| `risk.DefaultPolicy` | 생산 2(guardian_wiring · measure) + 시험 18 | 기본 정책에 필드 추가 |
| `measure.gridPolicy` | `Grid` · `Population` | 새 필수 칸 채움(격자는 기하만 거부해야 함) |

`codegraph affected <10 files> --filter '*_test.go'` → 1169 줄(`raw/affected-filter.txt`). 기본 판별식은 Go 시험을 못 찾으므로(.claude/CLAUDE.md) 필터 결과만 근거로 씀.
