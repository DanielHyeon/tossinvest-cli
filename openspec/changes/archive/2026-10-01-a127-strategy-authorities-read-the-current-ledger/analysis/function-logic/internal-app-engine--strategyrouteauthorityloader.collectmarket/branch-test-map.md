# Branch Test Map: `strategyRouteAuthorityLoader.collectMarket`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:148` `if !schedule.snapshot.Ready \ | schedule.restore.Activation == nil {` | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B2 | `:151` `if !candidates.snapshot.Ready {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B3 | `:154` `if candidates.approved.Len() == 0 {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B4 | `:157` `if loader.getenv == nil \ | loader.load == nil \ | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B5 | `:162` `if err != nil \ | base64.StdEncoding.EncodeToString(key) != encoded \ | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B6 | `:171` `for index := 0; index < candidates.approved.Len(); index++ {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B7 | `:173` `if !ok \ | !approved.Valid() \ | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B8 | `:187` `if err != nil \ | batch.ManifestDigest() != digest {` | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B9 | `:192` `for _, approved := range approvedValues {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B10 | `:194` `if !ok {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B11 | `:202` `if routed.Code != strategyrouter.RefusalNone \ | !routed.Valid() \ | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B12 | `:210` `if len(entries) == 0 {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
| B13 | `:215` `for _, entry := range entries {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S10b) | n/a | n/a |
