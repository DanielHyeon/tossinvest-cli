# Branch Test Map: `readProductionRiskUsage`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:535` `if err != nil {` | 아니오 | SQL 상수 이동(바이트 동일) — a126 · a127 시험군이 이 질의를 실행 | n/a | n/a |
| B2 | `:540` `for rows.Next() {` | 예 | SQL 상수 이동(바이트 동일) — a126 · a127 시험군이 이 질의를 실행 | n/a | n/a |
| B3 | `:542` `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | — | SQL 상수 이동(바이트 동일) — a126 · a127 시험군이 이 질의를 실행 | n/a | n/a |
