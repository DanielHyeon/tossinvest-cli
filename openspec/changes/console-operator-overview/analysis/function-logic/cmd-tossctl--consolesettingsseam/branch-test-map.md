# Branch Test Map: `consoleSettingsSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | seam이 만들어지지 않을 때 인터페이스에 typed-nil이 담기지 않는다 | `TestATypedNilSeamNeverReachesTheInterface` | — (동작 무변경) | yes |
