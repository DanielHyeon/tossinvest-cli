# Branch Test Map: `TestA091TheAugustSecondReplay`

- Source: `internal/app/engine/a091_replay_test.go`

> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.

| Branch | 조건 | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:109` `if res.sends != 1 \|\| res.pendingNew != 0 \|\| res.latched \|\| res.mode != journal.ModeNormal \|\| res.undeliveredLines != 0 {` | `TestA091TheAugustSecondReplay` | n/a | yes |
| B2 | `:117` `if res.pendingNew != 1 \|\| res.sends != 13 \|\| res.undeliveredLines != 1 \|\| !res.latched \|\| res.mode != journal.ModeEntryBlocked {` | `TestA091TheAugustSecondReplay` | n/a | yes |
| B3 | `:124` `if res.pendingNew != 1 \|\| !res.latched \|\| res.mode != journal.ModeEntryBlocked \|\| res.undeliveredLines != 14 \|\| res.noPublisherLines != 13 {` | `TestA091TheAugustSecondReplay` | n/a | yes |
| B4 | `:131` `if res.allRows != 0 \|\| res.sends != 0 \|\| res.latched \|\| res.mode != journal.ModeNormal \|\| res.undeliveredLines != 0 {` | `TestA091TheAugustSecondReplay` | n/a | yes |
