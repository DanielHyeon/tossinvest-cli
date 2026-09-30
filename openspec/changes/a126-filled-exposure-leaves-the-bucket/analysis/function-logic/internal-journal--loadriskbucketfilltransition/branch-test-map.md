# Branch Test Map: `loadRiskBucketFillTransition`

분기 · 동작 무편집(해소 조각을 상수로 — 같은 문자열). 기존 체결 시험 전부 통과(`analysis/impl/regress-1.log`), F3(조각 OR→AND) CAUGHT.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:768` `if err != nil \|\| cumulative == 0 {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B2 | `:773` `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND mar…` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B3 | `:779` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B4 | `:783` `for rows.Next() {` | 예 | 기존 — 무편집 | n/a | n/a |
| B5 | `:786` `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B6 | `:791` `if !isRiskBucketDimension(key.Dimension) {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B7 | `:795` `if decisionBuckets[decisionID] == nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B8 | `:798` `if decisionBuckets[decisionID][key] {` | 예 | 기존 — 무편집 | n/a | n/a |
| B9 | `:804` `if usage.LimitMinor == "" {` | 예 | 기존 — 무편집 | n/a | n/a |
| B10 | `:806` `} else {` | 예 | 기존 — 무편집 | n/a | n/a |
| B11 | `:809` `if currentErr != nil \|\| candidateErr != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B12 | `:813` `if candidateLimit.Cmp(currentLimit) < 0 {` | 예 | 기존 — 무편집 | n/a | n/a |
| B13 | `:819` `if addErr != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B14 | `:824` `if addErr != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B15 | `:829` `if addErr != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B16 | `:833` `if usage.Latches == nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B17 | `:840` `if err := rows.Close(); err != nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B18 | `:843` `for _, seen := range decisionBuckets {` | 예 | 기존 — 무편집 | n/a | n/a |
| B19 | `:844` `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B20 | `:848` `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {` | 예 | 기존 — 무편집 | n/a | n/a |
| B21 | `:854` `if err != nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B22 | `:859` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B23 | `:864` `for orders.Next() {` | 예 | 기존 — 무편집 | n/a | n/a |
| B24 | `:867` `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B25 | `:872` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B26 | `:877` `for key := range reserved {` | 예 | 기존 — 무편집 | n/a | n/a |
| B27 | `:880` `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {` | 예 | 기존 — 무편집 | n/a | n/a |
| B28 | `:888` `if err := orders.Close(); err != nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B29 | `:892` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B30 | `:895` `for fills.Next() {` | 예 | 기존 — 무편집 | n/a | n/a |
| B31 | `:899` `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B32 | `:904` `if orderIdentity == "" {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B33 | `:911` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B34 | `:915` `for alloc.Next() {` | 예 | 기존 — 무편집 | n/a | n/a |
| B35 | `:917` `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B36 | `:926` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B37 | `:933` `if err := alloc.Close(); err != nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B38 | `:940` `if err := fills.Close(); err != nil {` | 예 | 기존 — 무편집 | n/a | n/a |
| B39 | `:944` `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B40 | `:948` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
