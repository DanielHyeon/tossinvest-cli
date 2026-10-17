# Branch Test Map: `LoadProductionRouteAuthorityBatch`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:330` `if ctx == nil \ | config.ObservedAt.IsZero() {` | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B2 | `:333` `if err := ctx.Err(); err != nil {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B3 | `:337` `if config.JournalSchemaVersion <= 0 {` | 예 | `TestA127RouteLoaderRefusesAMissingInjectionBeforeOpeningTheLedger` — S6c · S6d | yes | yes |
| B4 | `:344` `if !ownerOK \ | name == "" \ | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B5 | `:352` `if err != nil \ | productionRouteDigest(data) != config.ManifestDigest {` | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B6 | `:356` `if err != nil \ | len(manifest.Scopes) == 0 {` | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B7 | `:362` `if _, verified := verifyProductionRouteManifest(manifest, verificationConfig); !verified {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B8 | `:366` `if err != nil {` | 예 | opener 원인 보존 — `TestA127RouteLoaderRefusesANewerLedgerAndSaysSoAtTheBatchBoundary` · `…OlderLedger…`(Batch 경계 문구) — S12a | yes | yes |
| B9 | `:380` `if err != nil \ | EvaluateMarketLifecycle(record, config.ObservedAt) != LifecycleReady {` | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B10 | `:384` `for _, target := range targets {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B11 | `:385` `if err := ctx.Err(); err != nil {` | 아니오 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B12 | `:389` `if !found {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B13 | `:393` `if err != nil {` | 아니오 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B14 | `:397` `if err != nil \ | revision != scope.OwnerRevision {` | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B15 | `:401` `if err != nil {` | 아니오 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B16 | `:405` `for _, value := range scope.Candidates {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
| B17 | `:423` `if err := tx.Commit(); err != nil {` | 예 | 기존 — production_test 시험군 · `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen`(실 원장, 외부 패키지) | n/a | n/a |
