# Branch Test Map: `LoadProductionRouteAuthorityBatch`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:319` `if ctx == nil \ | config.ObservedAt.IsZero() {` | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B2 | `:322` `if err := ctx.Err(); err != nil {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B3 | `:329` `if !ownerOK \ | name == "" \ | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B4 | `:337` `if err != nil \ | productionRouteDigest(data) != config.ManifestDigest {` | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B5 | `:341` `if err != nil \ | len(manifest.Scopes) == 0 {` | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B6 | `:347` `if _, verified := verifyProductionRouteManifest(manifest, verificationConfig); !verified {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B7 | `:351` `if err != nil {` | 아니오 | 편집 전 진입 0 — 픽스처가 `user_version=27` | n/a | n/a |
| B8 | `:364` `if err != nil \ | EvaluateMarketLifecycle(record, config.ObservedAt) != LifecycleReady {` | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B9 | `:368` `for _, target := range targets {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B10 | `:369` `if err := ctx.Err(); err != nil {` | 아니오 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B11 | `:373` `if !found {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B12 | `:377` `if err != nil {` | 아니오 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B13 | `:381` `if err != nil \ | revision != scope.OwnerRevision {` | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B14 | `:385` `if err != nil {` | 아니오 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B15 | `:389` `for _, value := range scope.Candidates {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
| B16 | `:407` `if err := tx.Commit(); err != nil {` | 예 | 기존 — `strategyrouter/production_test.go` 시험군 | n/a | n/a |
