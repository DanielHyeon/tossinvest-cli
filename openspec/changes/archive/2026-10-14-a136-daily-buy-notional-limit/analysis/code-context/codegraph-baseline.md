# a136 CodeGraph baseline

- 도구: codegraph 1.6.0, 질의 시점 HEAD `45e424d1`(코드 = base 7b073f55 와 동일 — 그 사이 커밋은 문서뿐). 원문은 `raw/`.

| 심볼 | 생산 호출자 (callers) | a136 영향 |
|---|---|---|
| `journal.reserveRows` | `RecordDecisionAndReserve`(issuance.go:106) · `Reserve`(reservations.go:249) · first-leg atomic(strategy_first_leg_atomic.go:101) · strategy lineage(strategy_lineage.go:315) · risk-bucket issuance(risk_bucket_issuance.go:70) | **모든 예약 tx 가 이 한 함수를 지난다** → 판정 훅 한 자리 |
| `RiskGuardian.IssueEntry` | tracer.go:414 · `IssueStrategyEntry`(→ strategydispatch adapters.go:41) | 발급 문 1 |
| `IssuePrecheckedQFinalEntry` | `IssueQFinalEntry`(생산 호출자 0) | 발급 문 2 |
| `IssuePrecheckedQFinalCampaignFirstLeg` | strategy_first_leg_admission.go:105 | 발급 문 3(생산 전략 경로) |
| `riskPolicyFromAutomationGate` | `newProductionRiskGuardian` | 정책 유도 한 곳 |
| `verifyGate` · `acceptanceDetail` | `runInterlock`(interlock.go:376) | 인터록 대조 · 감사 문구 |
| `newAutomationStatus` | engine.go:437 | 해석 값 운반 |
| `checkGuardianLimits` | `verifyGate`(interlock.go:478 정의 부근) | 무편집(새 함수를 옆에 둠) |
| `mergeEngine` | `Service` 읽기(service.go:240) | 설정 필드 복사 |
| `risk.DefaultPolicy` | guardian_wiring.go:34 · measure/counterfactual.go:353 · 시험 다수 | 기본 정책에 필드 추가 |
| `sweepTradingDayBoundary` | `sweepLapsedReservations` | 무편집(근거만) |

`codegraph affected <8 files> --filter '*_test.go'` → 1144 줄(`raw/affected-filter.txt`). 기본 판별식은 Go 시험을 못 찾으므로(.claude/CLAUDE.md) 필터 결과만 근거로 씀.
