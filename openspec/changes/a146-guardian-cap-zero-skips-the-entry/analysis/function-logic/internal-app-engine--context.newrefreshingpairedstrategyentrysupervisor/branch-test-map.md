# Branch Test Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (414-443)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 415 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B2 | if 419 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B3 | range 423 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B4 | if 436 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
