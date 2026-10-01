# Branch Test Map: `LoadProductionRiskSnapshotAuthority`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:171` `if ctx == nil \ | config.ObservedAt.IsZero() {` | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B2 | `:174` `if err := ctx.Err(); err != nil {` | 예 | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B3 | `:178` `if config.JournalSchemaVersion <= 0 {` | 예 | `TestA127RiskLoaderRefusesAMissingInjectionBeforeOpeningTheLedger`(0 · 음수, 존재하지 않는 원장 경로) — 변이 S6a · S6b | yes | yes |
| B4 | `:184` `if !ownerOK \ | name == "" \ | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B5 | `:190` `if err != nil \ | productionRiskDigest(data) != config.ManifestDigest {` | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B6 | `:194` `if err != nil \ | !verifyProductionRiskPolicy(manifest, config) {` | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B7 | `:199` `if err != nil {` | 예 | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
| B8 | `:203` `if err != nil {` | 예 | 기존 — production_snapshot_authority_test 시험군 · `TestA127RiskLoaderReadsTheLedgerWhoseSchemaEqualsTheInjectedVersion` | n/a | n/a |
