# Branch Test Map: `loadProductionRouteOwnersFrom`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:625` `if queryer == nil {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B2 | `:629` `if err != nil {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B3 | `:634` `for rows.Next() {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B4 | `:635` `if len(history) >= productionRouteMaxOwners {` | 아니오 | 기존 — production_test 시험군 | n/a | n/a |
| B5 | `:639` `if err := rows.Scan(&value.prospective, &value.laneID, &value.campaignID, &value.actual, &value.acquired, &value.released, &value.overage…` | — | 기존 — production_test 시험군 | n/a | n/a |
| B6 | `:645` `if err := rows.Err(); err != nil {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B7 | `:651` `for _, value := range history {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B8 | `:652` `if value.released == "" {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B9 | `:656` `if len(active) > 1 {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B10 | `:659` `if len(active) == 0 {` | 예 | 기존 — production_test 시험군 | n/a | n/a |
| B11 | `:663` `if value.actual == "" \ | value.overage != 0 \ | 기존 — production_test 시험군 | n/a | n/a |
| B12 | `:667` `if err != nil \ | actual != key.PositionGeneration \ | 기존 — production_test 시험군 | n/a | n/a |
| B13 | `:672` `if err := queryer.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_version,prospective_token,coalesce(actual_position_generati…` | — | 기존 — production_test 시험군 | n/a | n/a |
| B14 | `:679` `if !ok \ | descriptor.LaneVersion != laneVersion {` | 기존 — production_test 시험군 | n/a | n/a |
