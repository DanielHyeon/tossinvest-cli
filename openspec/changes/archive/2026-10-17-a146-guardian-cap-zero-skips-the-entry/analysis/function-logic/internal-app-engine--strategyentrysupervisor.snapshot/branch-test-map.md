# Branch Test Map: `Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (797-820)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 798 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B2 | if 804 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
