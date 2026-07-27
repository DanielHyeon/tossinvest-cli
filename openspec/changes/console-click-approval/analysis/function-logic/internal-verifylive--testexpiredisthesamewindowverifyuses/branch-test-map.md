# Branch Test Map: `TestExpiredIsTheSameWindowVerifyUses`

- Source: `internal/verifylive/confirm_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if b.Expired(now) {` (internal/verifylive/confirm_test.go:261) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if !b.Expired(after) {` (internal/verifylive/confirm_test.go:265) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if err := b.Verify(b.Nonce, after); !errors.Is(err, ErrConfirmationExpired) {` (internal/verifylive/confirm_test.go:268) | 이 함수 자체가 테스트다 | yes | yes |
