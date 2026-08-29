# Branch Test Map: `TestTheNonceIsJudgedByVerifylive`

- Source: `internal/console/static_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !strings.Contains(code, "view.Batch.Verify(") {` (internal/console/static_test.go:200) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, banned := range []string{"== view.Batch.Nonce", "Nonce ==", "== nonce"} {` (internal/console/static_test.go:203) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if strings.Contains(code, banned) {` (internal/console/static_test.go:204) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `for name, fileSrc := range packageFiles(t) {` (internal/console/static_test.go:209) | 이 함수 자체가 테스트다 | yes | yes |
| B5 | `for _, banned := range []string{"AUTO_APPROVE", "SKIP_CONFIRM", "NO_CONFIRM", "TOSSCTL_CONSOLE_TOKEN"} {` (internal/console/static_test.go:211) | 이 함수 자체가 테스트다 | yes | yes |
| B6 | `if strings.Contains(body, banned) {` (internal/console/static_test.go:212) | 이 함수 자체가 테스트다 | yes | yes |
