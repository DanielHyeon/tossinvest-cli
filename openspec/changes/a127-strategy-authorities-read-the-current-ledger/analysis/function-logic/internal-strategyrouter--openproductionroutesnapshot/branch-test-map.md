# Branch Test Map: `openProductionRouteSnapshot`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:601` `if err := validateProductionRouteJournalFile(journalPath, ownerUID); err != nil {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B2 | `:606` `if err != nil {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B3 | `:611` `if err != nil {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B4 | `:616` `if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil \ | version != productionRouteJournalV {` | 편집 전 진입 0 — 픽스처가 `user_version=27` | n/a | n/a |
