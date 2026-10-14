# a137 CodeGraph baseline

- 도구: codegraph **1.6.0**(`codegraph --version`), 기준 HEAD `7b073f55`(base 미고정 — 사람 결정 · a136 착지 뒤 고정), 원문은 `raw/`.
- 질의와 결론:

| 질의 | 결과(raw) | 결론 |
|---|---|---|
| `callers DailyLoss` | 6 — 전부 `riskcalc_test.go` | `riskcalc.DailyLoss` 생산 호출자 0 |
| `callers checkDailyLoss` | `entryChain`(chain.go:96) | 진입 체인 단 하나로만 도달 |
| `impact checkDailyLoss` | 5 — checkDailyLoss · entryChain · EntryChainSteps · Evaluate · chain.go | 영향은 risk 패키지 체인 안 |
| `callers escalateFor` | 1 — `IssueEntry` | 자동 강화는 레거시 진입 입구에만 |
| `callers IssueEntry` | 12 — 생산: `Tracer.submitEntry`(tracer.go:406) · `IssueStrategyEntry`(riskguardian.go:339), 나머지 시험 | 두 생산 입구 |
| `callers PrecheckQFinalEntry` | 2 — `PrecheckQFinalCampaignFirstLeg` · `IssueQFinalEntry` | 생산 전략 첫 다리가 이 입구로 체인을 돈다 |
| `callers PrecheckQFinalCampaignFirstLeg` | 1 — 인터페이스 메서드(strategy_first_leg_admission.go:56) | 실호출은 같은 파일 :91(HEAD grep) |
| `callers IssueStrategyEntry` | 1 — `GuardianAdapter.IssueAndPlan` | `GuardianAdapter` 생산 생성자 0(HEAD grep) |
| `callers LoadProductionAuthority` | 2 — 시험만 | **불일치** — 생산은 함수 값 참조(`load: strategyaccount.LoadProductionAuthority`, strategy_account_first_leg_authority.go:116)라 CodeGraph 가 호출로 세지 않음. evidence-reconciliation R4 |
| `callers accountState` | 1 — `Tracer.submitEntry` | tracer 리터럴 0 의 소비자 하나 |
| `callers submitEntry` | 1 — `Tracer.Run` | `NewTracer` 생산 호출자 0(HEAD grep) |
| `callers TradeOutcomes` | 1 — 시험 | 계좌 단위 성과 읽기의 생산 소비자 없음(콘솔은 `ReadOnly` 경로) |
| `affected chain.go riskguardian.go riskguardian_qfinal.go --filter '*_test.go'` | 1131 시험 파일 | 기본 판별식은 Go 시험을 못 찾음(.claude/CLAUDE.md 하네스 주석) — `--filter` 로 잼 |
