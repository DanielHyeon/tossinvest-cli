# Branch Test Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if err != nil {` (cmd/tossctl/console.go:365) | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) | yes | yes |
| B2 | `if err != nil {` (cmd/tossctl/console.go:369) | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) | yes | yes |
| B3 | `if err != nil {` (cmd/tossctl/console.go:374) | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) | yes | yes |
| B4 | `if err != nil {` (cmd/tossctl/console.go:395) | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) | yes | yes |
| B5 | `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` (cmd/tossctl/console.go:404) | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) | yes | yes |
