# Branch Test Map: `loadProductionRiskEntries`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:392` `if !ok {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B2 | `:395` `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | 예 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B3 | `:404` `if err != nil {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B4 | `:409` `if err := db.PingContext(ctx); err != nil {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B5 | `:415` `if err != nil {` | 아니오 | BeginTx — 구조 단언 `TestA127RiskLoaderReadsEverythingInOneReadOnlyTransaction`(S13a~c) | yes | yes |
| B6 | `:422` `if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {` | 아니오 | 버전 판독 — 같은 tx 구조 단언(S13c) | yes | yes |
| B7 | `:425` `if version > config.JournalSchemaVersion {` | 예 | `TestA127RiskLoaderRefusesANewerLedger`(방향 문구 · ScopeRefused 아님) — S3 · S9a · S11 | yes | yes |
| B8 | `:428` `if version < config.JournalSchemaVersion {` | 예 | `TestA127RiskLoaderRefusesAnOlderLedger` — S5a; 양성 `TestTheRiskLoaderReadsTheRealJournal`(engine, 실 원장) — S1 | yes | yes |
| B9 | `:433` `for _, statement := range []string{productionRiskScopeLatchSQL, productionRiskUsageSQL} {` | 예 | prepare 선행 — `TestA127RiskLoaderReportsAMissingUsageColumnAsADefectEvenOnALatchedScope` — S14 | yes | yes |
| B10 | `:435` `if err != nil {` | 예 | 같은 시험(사용량 전용 열 삭제 · latch 범위 → 결함 신원) | yes | yes |
| B11 | `:442` `if err := tx.QueryRowContext(ctx, productionRiskScopeLatchSQL, scope.AccountID, string(scope.Market), scope.Symbol).Scan(&scopeLatches); …` | 아니오 | scope latch 판독 — 같은 tx 구조 단언(S13b) | yes | yes |
| B12 | `:445` `if scopeLatches != 0 {` | 예 | 같은 시험의 대조(온전한 원장 · latch → ScopeRefused) | n/a | n/a |
| B13 | `:455` `if authorityObserved.After(scope.AsOf) \ | authorityFresh.Before(scope.AsOf) {` | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B14 | `:459` `for _, dimension := range requiredDimensions {` | 예 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B15 | `:461` `if err != nil {` | 아니오 | 사용량 판독 — 같은 tx 구조 단언(S13a) | yes | yes |
| B16 | `:465` `if usage.Latched {` | 예 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B17 | `:475` `if err != nil {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B18 | `:485` `if err != nil {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
| B19 | `:494` `if err != nil {` | 아니오 | 기존 — 이 change 가 판정을 바꾸지 않은 분기 | n/a | n/a |
