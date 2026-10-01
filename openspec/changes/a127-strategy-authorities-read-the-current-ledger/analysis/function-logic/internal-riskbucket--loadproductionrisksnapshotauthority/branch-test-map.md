# Branch Test Map: `LoadProductionRiskSnapshotAuthority`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:146` `if ctx == nil \ | config.ObservedAt.IsZero() {` | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B2 | `:149` `if err := ctx.Err(); err != nil {` | 예 | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B3 | `:155` `if !ownerOK \ | name == "" \ | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B4 | `:161` `if err != nil \ | productionRiskDigest(data) != config.ManifestDigest {` | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B5 | `:165` `if err != nil \ | !verifyProductionRiskPolicy(manifest, config) {` | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B6 | `:170` `if err != nil {` | 예 | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
| B7 | `:174` `if err != nil {` | 예 | 기존 — `production_snapshot_authority_test.go`(tossos_testseams) 시험군 | n/a | n/a |
