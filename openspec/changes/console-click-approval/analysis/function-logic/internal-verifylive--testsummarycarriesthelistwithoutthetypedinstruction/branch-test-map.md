# Branch Test Map: `TestSummaryCarriesTheListWithoutTheTypedInstruction`

- Source: `internal/verifylive/confirm_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, want := range []string{` (internal/verifylive/confirm_test.go:226) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !strings.Contains(summary, want) {` (internal/verifylive/confirm_test.go:229) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `for _, banned := range []string{b.Nonce, "확인 문자열", "입력하라"} {` (internal/verifylive/confirm_test.go:233) | 이 함수 자체가 테스트다 | yes | yes |
| B4 | `if strings.Contains(summary, banned) {` (internal/verifylive/confirm_test.go:234) | 이 함수 자체가 테스트다 | yes | yes |
