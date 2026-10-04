# Branch Test Map: `LoadProductionRiskSnapshotAuthority`

- **a112 재추출(2026-10-04 게이트 준비).** a127 `82080177`(전략 권한 적재기가 현재 원장을 읽음)이 이 함수의 분기 구조를 바꿔 a112 번들이 낡았다 — 현재 AST 와 같은 a127 아카이브 번들(`openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/function-logic/internal-riskbucket--loadproductionrisksnapshotauthority/`)을 옮겨 왔다(그 판의 RED · 변이 경로는 아카이브 좌표로 고쳐 씀). a127 이전 a112 판(a112 의 편집 기록)은 `git show cc79c887:openspec/changes/a112-run-four-strategy-families-independently/analysis/function-logic/internal-riskbucket--loadproductionrisksnapshotauthority/function-logic-map.md` · `branch-test-map.md`.

편집 뒤 — RED(`openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/impl/red.log`) → GREEN → 변이(`openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

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
