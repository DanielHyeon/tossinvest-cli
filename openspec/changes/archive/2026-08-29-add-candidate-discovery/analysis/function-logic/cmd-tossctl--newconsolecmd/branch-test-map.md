# Branch Test Map: `newConsoleCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 명령이 등록되고 `mutating=true`·`source=official`이며 플래그는 `--port` 하나뿐, 도움말이 화면 목록을 정확히 적는다 | `TestConsoleIsRegisteredAndAnnotated` + `TestConsoleOffersOnlyThePortFlag` | yes (도움말 화면 목록 변경 시 FAIL 없음 — 플래그 수 가드는 RED 재현) | yes |
