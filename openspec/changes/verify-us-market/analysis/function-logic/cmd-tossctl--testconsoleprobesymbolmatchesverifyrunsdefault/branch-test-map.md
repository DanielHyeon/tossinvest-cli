# Branch Test Map: `TestConsoleProbeSymbolMatchesVerifyRunsDefault`

- Source: `cmd/tossctl/console_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if run == nil {` (cmd/tossctl/console_test.go:131) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if f == nil {` (cmd/tossctl/console_test.go:135) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if f.DefValue != consoleProbeSymbolKR {` (cmd/tossctl/console_test.go:138) | 이 함수 자체가 테스트다 | yes | yes |
