# Branch Test Map: `New`

- Source: `internal/verifylive/runner.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if o.Broker == nil {` (internal/verifylive/runner.go:154) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B2 | `if o.Recorder == nil {` (internal/verifylive/runner.go:157) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B3 | `if o.Confirm == nil {` (internal/verifylive/runner.go:160) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B4 | `if o.ConfirmBatch == nil {` (internal/verifylive/runner.go:163) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B5 | `if strings.TrimSpace(o.AccountRef) == "" {` (internal/verifylive/runner.go:167) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B6 | `if err != nil {` (internal/verifylive/runner.go:171) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B7 | `if r.approvalChannel == "" {` (internal/verifylive/runner.go:195) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B8 | `if r.out == nil {` (internal/verifylive/runner.go:198) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B9 | `if r.now == nil {` (internal/verifylive/runner.go:201) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B10 | `if r.sleep == nil {` (internal/verifylive/runner.go:204) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B11 | `if r.maxSellQuantity <= 0 {` (internal/verifylive/runner.go:207) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B12 | `if r.ttlWait <= 0 {` (internal/verifylive/runner.go:210) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B13 | `if r.process.InstanceID == "" {` (internal/verifylive/runner.go:213) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
| B14 | `for _, id := range o.Redo {` (internal/verifylive/runner.go:216) | TestTheApprovalChannelIsRecordedAsGiven | yes | yes |
