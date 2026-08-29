# Branch Test Map: `TestPromptIsTheSummaryPlusTheTypedTail`

- Source: `internal/verifylive/confirm_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if !strings.HasPrefix(b.Prompt(), b.Summary()) {` (internal/verifylive/confirm_test.go:245) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `for _, want := range []string{b.Nonce, "확인 문자열", "입력하라"} {` (internal/verifylive/confirm_test.go:249) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !strings.Contains(tail, want) {` (internal/verifylive/confirm_test.go:250) | 이 함수 자체가 테스트다 | yes | yes |
