# Branch Test Map: `readProductionRiskUsage`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:477` `if err != nil {` | 아니오 | 기존 — 무편집 분기 | n/a | n/a |
| B2 | `:482` `for rows.Next() {` | 예 | `TestA126AReleasedOwnersFilledLeavesEveryBucket` · `TestA126CorruptReceiptedRowsAreUnreadable` · `TestA126ALateBuyViaRecordFillWithCampaignHookRevertsTheDeparture` — 변이 M9 · M10 CAUGHT | yes | yes |
| B3 | `:484` `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | — | 스캔 — 사실 열 6 을 싣는 자리(M9 · M10 이 이 열들을 바꿔치기해 잼) | yes | yes |
