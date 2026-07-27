# Branch Test Map: `Batch.Verify`

- Source: `internal/verifylive/confirm.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if b.Expired(now) {` (internal/verifylive/confirm.go:254) | TestConfirmBatchRejectsAnythingButTheNonce, TestConfirmBatchExpires, TestExpiredIsTheSameWindowVerifyUses | yes | yes |
| B2 | `if strings.TrimSpace(input) != b.Nonce {` (internal/verifylive/confirm.go:257) | TestConfirmBatchRejectsAnythingButTheNonce, TestConfirmBatchExpires, TestExpiredIsTheSameWindowVerifyUses | yes | yes |
