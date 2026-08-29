# Branch Test Map: `assertNothingWasSent`

- Source: `internal/verifylive/plan_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, req := range broker.seen() {` (internal/verifylive/plan_test.go:791) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if strings.HasPrefix(req, "POST ") \|\| strings.HasPrefix(req, "DELETE ") {` (internal/verifylive/plan_test.go:792) | 이 함수 자체가 테스트다 | yes | yes |
