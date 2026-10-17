# Branch Test Map: `TestA091TheProductionAssemblyReadsTheLoadedSwitch`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go`

> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.

| Branch | 조건 | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:873` `for _, loaded := range []bool{true, false} {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
| B2 | `:882` `if err != nil {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
| B3 | `:885` `if eng.Log == nil {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
| B4 | `:890` `if err != nil {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
| B5 | `:893` `if got := observer.OptionsForTest().ZeroFloorLog; got == nil \|\| got != eng.Log {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
| B6 | `:896` `if got := observer.OptionsForTest().NotificationsEnabled; got != loaded {` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` | n/a | yes |
