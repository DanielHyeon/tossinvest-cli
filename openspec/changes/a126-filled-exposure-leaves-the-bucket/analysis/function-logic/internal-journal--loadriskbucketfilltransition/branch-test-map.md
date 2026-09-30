# Branch Test Map: `loadRiskBucketFillTransition`

분기 · 동작 무편집(해소 조각을 상수로 — 같은 문자열). 기존 체결 시험이 전부 그대로 통과해야 한다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:762` `if err != nil \|\| cumulative == 0 {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B2 | `:767` `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND mar…` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B3 | `:773` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B4 | `:777` `for rows.Next() {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B5 | `:780` `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B6 | `:785` `if !isRiskBucketDimension(key.Dimension) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B7 | `:789` `if decisionBuckets[decisionID] == nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B8 | `:792` `if decisionBuckets[decisionID][key] {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B9 | `:798` `if usage.LimitMinor == "" {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B10 | `:800` `} else {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B11 | `:803` `if currentErr != nil \|\| candidateErr != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B12 | `:807` `if candidateLimit.Cmp(currentLimit) < 0 {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B13 | `:813` `if addErr != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B14 | `:818` `if addErr != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B15 | `:823` `if addErr != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B16 | `:827` `if usage.Latches == nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B17 | `:834` `if err := rows.Close(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B18 | `:837` `for _, seen := range decisionBuckets {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B19 | `:838` `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B20 | `:842` `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B21 | `:848` `if err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B22 | `:853` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B23 | `:858` `for orders.Next() {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B24 | `:861` `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B25 | `:866` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B26 | `:871` `for key := range reserved {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B27 | `:874` `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B28 | `:882` `if err := orders.Close(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B29 | `:886` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B30 | `:889` `for fills.Next() {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B31 | `:893` `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B32 | `:898` `if orderIdentity == "" {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B33 | `:905` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B34 | `:909` `for alloc.Next() {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B35 | `:911` `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B36 | `:920` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B37 | `:927` `if err := alloc.Close(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B38 | `:934` `if err := fills.Close(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B39 | `:938` `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B40 | `:942` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
