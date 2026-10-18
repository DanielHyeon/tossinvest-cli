# Branch Test Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (325-413)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 326 | 조립 불가 오류(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B2 | if 332 | 원장 경로 유도(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B3 | if 343 | 증거 DB 경로 유도(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B4 | if 353 | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |
| B5 | if 374 | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock` |
| B6 | range 390 | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B7 | if 398 | **a157** 게이트가 있을 때만 배선 — 없으면 인터페이스 칸이 nil 로 남음(F10) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestA157TheAssemblySupervisorEntryGate` |
| B8 | if 402 | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |
| B9 | if 409 | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |
