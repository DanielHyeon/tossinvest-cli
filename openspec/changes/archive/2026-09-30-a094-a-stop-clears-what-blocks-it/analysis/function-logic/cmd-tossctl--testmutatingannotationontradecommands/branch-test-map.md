# Branch Test Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:147` `for _, c := range leafCommands(newRootCmd()) {` | — | `TestMutatingAnnotationOnTradeCommands` | n/a | yes |
| B2 | `:150` `if wantMutating[path] && !isMut {` | — | `TestMutatingAnnotationOnTradeCommands` | n/a | yes |
| B3 | `:153` `if !wantMutating[path] && isMut {` | — | `TestMutatingAnnotationOnTradeCommands` | n/a | yes |
