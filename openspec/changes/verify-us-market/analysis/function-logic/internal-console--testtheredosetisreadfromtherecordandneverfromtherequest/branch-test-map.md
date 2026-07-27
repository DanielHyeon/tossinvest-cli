# Branch Test Map: `TestTheRedoSetIsReadFromTheRecordAndNeverFromTheRequest`

- Source: `internal/console/static_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !strings.Contains(code, "c.redoSet(market)") {` (internal/console/static_test.go:235) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, banned := range []string{` (internal/console/static_test.go:238) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if strings.Contains(code, banned) {` (internal/console/static_test.go:242) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if !strings.Contains(data, "verifylive.RedoSet(") {` (internal/console/static_test.go:248) | 이 함수 자체가 테스트다 | yes | yes |
