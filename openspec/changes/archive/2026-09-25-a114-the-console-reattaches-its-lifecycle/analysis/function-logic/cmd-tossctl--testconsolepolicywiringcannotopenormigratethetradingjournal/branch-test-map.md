# Branch Test Map: `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 금지 문자열 순회 | `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal` | no — 측정 안 함(GREEN 첫 실행의 -run 필터 밖이었다; console.go 만 읽는 판본은 `positionpolicyrpc.Dial` 필수 문자열에서 실패했을 것이나 관측하지 않았다) | yes |
| B2 | 금지 문자열 발견 | `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal` | no | yes |
| B3 | 필수 좁은 client 문자열 순회 | `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal` | no | yes |
| B4 | 필수 문자열 없음 | `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal` | no | yes |
