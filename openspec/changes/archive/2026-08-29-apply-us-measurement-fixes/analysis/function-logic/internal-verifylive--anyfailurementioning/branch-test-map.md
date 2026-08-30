# Branch Test Map: `anyFailureMentioning`

- Source: `internal/verifylive/fake_broker_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, e := range entries {` (internal/verifylive/fake_broker_test.go:870) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if e.Verdict == VerdictFail && strings.Contains(e.Reason, text) {` (internal/verifylive/fake_broker_test.go:871) | 이 함수 자체가 테스트다 | yes | yes |
