# a146 CodeGraph baseline

- 도구: codegraph 1.6.0, 질의 시점 HEAD `befea6ee`(`codegraph sync .` 직후). 원문은 `raw/`.

| 심볼 | 생산 호출자 (CodeGraph) | a146 영향 |
|---|---|---|
| `engine.strategyFirstLegAdmissionBridge.admit` | 이름 충돌로 다른 `admit`(strategyworker · coordinator)과 섞임 — `strategyDispatchCycle.dispatch` 가 목록에 없음(reconciliation) | **편집** — 분류 갈래 하나 |
| `engine.deliverEachStrategyHandoff` | `dispatchStrategyMarketHandoffs` 하나 + 시험 7 | **편집** — 스킵 `continue` |
| `engine.dispatchStrategyMarketHandoffs` | `runProductionStrategyMarketCycle`(HEAD grep) | 무편집 |
| `engine.StrategyEntrySupervisor.runMarket` | `Run` 하나 | **편집** — 기록자 한 줄 |
| `engine.StrategyEntrySupervisor.latchMarket` | `runMarket` 하나 | 무편집 |
| `engine.NewStrategyEntrySupervisor` | 생산 3(`NewPairedStrategyEntryProductionAssembly` 둘 · `NewRefreshingPairedStrategyEntrySupervisor`) + 시험 | **편집** — 옵션 칸 하나 |
| `engine.Context.NewRefreshingPairedStrategyEntrySupervisor` | `cmd/tossctl.engineRuntime` 하나 + 시험 4 | **편집** — 알림 배선 한 줄 |
| `engine.StrategyEntrySupervisor.Snapshot` | 이름 충돌 다수 — 생산 소비자는 `strategyProjectionFromAssembly`(HEAD grep) | **편집** — 칸 둘 |
| `execgw.RiskGuardian.PrecheckQFinalEntry` · `PrecheckQFinalCampaignFirstLeg` | `PrecheckQFinalCampaignFirstLeg` / `admit`(인터페이스) | 무편집 |
| `risk.AccountBaseStrategyEntryQuantity` · `StrategyEntryQuantity` | `PrecheckQFinalEntry`(+ `IssueStrategyEntry`) | 무편집 |

`codegraph affected <admission · delivery · supervisor 3 files> --filter '*_test.go'` → `raw/affected-filter.txt`(1,149 시험 파일 — 기본 판별식은 Go 시험을 못 찾음,
.claude/CLAUDE.md 하네스 주석). 대상 시험 패키지는 `internal/app/engine`(tossos_testseams 포함) · `cmd/tossctl`(감독자 생성).
