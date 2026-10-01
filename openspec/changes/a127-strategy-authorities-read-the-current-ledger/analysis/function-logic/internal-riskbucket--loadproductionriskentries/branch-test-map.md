# Branch Test Map: `loadProductionRiskEntries`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:363` `if !ok {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B2 | `:366` `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | 예 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B3 | `:375` `if err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B4 | `:380` `if err := db.PingContext(ctx); err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B5 | `:384` `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \ | version != productionRiskJournalSchema {` | 편집 전 진입 0 — 모든 픽스처가 `user_version=27`(가린 픽스처, proposal) | n/a | n/a |
| B6 | `:389` `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` | — | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B7 | `:393` `if scopeLatches != 0 {` | 예 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B8 | `:403` `if authorityObserved.After(scope.AsOf) \ | authorityFresh.Before(scope.AsOf) {` | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B9 | `:407` `for _, dimension := range requiredDimensions {` | 예 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B10 | `:409` `if err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B11 | `:413` `if usage.Latched {` | 예 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B12 | `:423` `if err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B13 | `:433` `if err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
| B14 | `:442` `if err != nil {` | 아니오 | 기존 — production_snapshot_authority_test · a126 replay 시험 | n/a | n/a |
