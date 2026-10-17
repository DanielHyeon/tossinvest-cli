# Branch Test Map: `strategyRiskAuthorityLoader.collectMarket`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:194` `if !result.ready {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B2 | `:197` `if !fx.snapshot.Ready \ | !fx.read.valid {` | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B3 | `:201` `if market == StrategyMarketUS {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B4 | `:207` `for _, scoped := range result.results() {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B5 | `:211` `if keyed {` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B6 | `:220` `switch {` | — | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B7 | `:221` `case err != nil:` | 예 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B8 | `:224` `case string(scope.Market) == string(market) && scope.AccountID == loader.accountID &&` | — | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
| B9 | `:227` `default:` | 아니오 | config 필드 하나 — 구조 단언 `TestA127EngineInjectsTheJournalSchemaVersionConstant`(S7 · S10a), 실 원장 `TestTheRiskLoaderReadsTheRealJournal` | n/a | n/a |
