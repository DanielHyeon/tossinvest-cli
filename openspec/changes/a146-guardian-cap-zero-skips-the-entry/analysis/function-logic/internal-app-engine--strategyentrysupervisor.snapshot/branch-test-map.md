# Branch Test Map: `Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (782-804)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 783 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B2 | if 789 | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
