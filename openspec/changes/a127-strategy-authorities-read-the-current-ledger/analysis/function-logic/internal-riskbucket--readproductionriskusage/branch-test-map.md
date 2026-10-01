# Branch Test Map: `readProductionRiskUsage`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:498` `if err != nil {` | 아니오 | 기존 — a126 시험군(실 원장) | n/a | n/a |
| B2 | `:503` `for rows.Next() {` | 예 | 기존 — a126 시험군(실 원장) | n/a | n/a |
| B3 | `:505` `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | — | 기존 — a126 시험군(실 원장) | n/a | n/a |
