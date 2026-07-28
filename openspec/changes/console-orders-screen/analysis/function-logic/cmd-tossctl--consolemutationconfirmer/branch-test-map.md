# Branch Test Map: `consoleMutationConfirmer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) confirmer가 항상 `ErrNotATerminal`로 거절하고, `verify run`은 여전히 터미널 confirmer만 쓴다 | `TestConsoleMutationConfirmerRefuses` + `TestVerifyRunStillConfirmsAtTheTerminalOnly` | — (동작 무변경) | yes |
