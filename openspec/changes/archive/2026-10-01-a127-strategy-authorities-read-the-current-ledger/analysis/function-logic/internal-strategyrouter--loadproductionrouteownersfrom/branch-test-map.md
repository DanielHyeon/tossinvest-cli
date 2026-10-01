# Branch Test Map: `loadProductionRouteOwnersFrom`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:661` `if queryer == nil {` | 아니오 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B2 | `:665` `if err != nil {` | 아니오 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B3 | `:670` `for rows.Next() {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B4 | `:671` `if len(history) >= productionRouteMaxOwners {` | 아니오 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B5 | `:675` `if err := rows.Scan(&value.prospective, &value.laneID, &value.campaignID, &value.actual, &value.acquired, &value.released, &value.overage…` | — | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B6 | `:681` `if err := rows.Err(); err != nil {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B7 | `:687` `for _, value := range history {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B8 | `:688` `if value.released == "" {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B9 | `:692` `if len(active) > 1 {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B10 | `:695` `if len(active) == 0 {` | 예 | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B11 | `:699` `if value.actual == "" \ | value.overage != 0 \ | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B12 | `:703` `if err != nil \ | actual != key.PositionGeneration \ | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B13 | `:708` `if err := queryer.QueryRowContext(ctx, productionRouteCampaignSQL, value.campaignID, value.laneID).` | — | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
| B14 | `:715` `if !ok \ | descriptor.LaneVersion != laneVersion {` | SQL 상수 이동(바이트 동일) — production_test · a127 시험군 | n/a | n/a |
