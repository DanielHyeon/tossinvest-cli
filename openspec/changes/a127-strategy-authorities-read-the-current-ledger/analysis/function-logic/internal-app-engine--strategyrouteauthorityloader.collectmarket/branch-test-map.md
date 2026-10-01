# Branch Test Map: `strategyRouteAuthorityLoader.collectMarket`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:147` `if !schedule.snapshot.Ready \ | schedule.restore.Activation == nil {` | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B2 | `:150` `if !candidates.snapshot.Ready {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B3 | `:153` `if candidates.approved.Len() == 0 {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B4 | `:156` `if loader.getenv == nil \ | loader.load == nil \ | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B5 | `:161` `if err != nil \ | base64.StdEncoding.EncodeToString(key) != encoded \ | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B6 | `:170` `for index := 0; index < candidates.approved.Len(); index++ {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B7 | `:172` `if !ok \ | !approved.Valid() \ | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B8 | `:184` `if err != nil \ | batch.ManifestDigest() != digest {` | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B9 | `:189` `for _, approved := range approvedValues {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B10 | `:191` `if !ok {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B11 | `:199` `if routed.Code != strategyrouter.RefusalNone \ | !routed.Valid() \ | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B12 | `:207` `if len(entries) == 0 {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
| B13 | `:212` `for _, entry := range entries {` | 예 | 기존 — strategy_route_authority_test 시험군(loader.load stub) | n/a | n/a |
