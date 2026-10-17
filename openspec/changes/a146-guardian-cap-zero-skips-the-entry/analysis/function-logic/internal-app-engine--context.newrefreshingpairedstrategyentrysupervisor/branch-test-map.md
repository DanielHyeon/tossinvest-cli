# Branch Test Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (399-428)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 400 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B2 | if 404 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B3 | range 408 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B4 | if 421 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
