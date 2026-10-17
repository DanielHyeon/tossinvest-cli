# Branch Test Map: `Journal.LiveOrdersForSymbol`

- Source: `internal/journal/fills.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1851` `if err := j.guardTrackedFillIdentity(ctx, accountRef); err != nil {` | 예 | `TestA094AnAmbiguouslyOwnedOrderIsStillAwaitingClose` | n/a | yes |
| B2 | `:1881` `if err != nil {` | 아니오 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B3 | `:1887` `for rows.Next() {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B4 | `:1889` `if err := rows.Scan(&o.OrderID, &o.IntentID, &o.AccountRef, &o.Market, &o.TradingDay,` | — | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B5 | `:1896` `if err := rows.Err(); err != nil {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B6 | `:1900` `for i := range out {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B7 | `:1908` `if err != nil {` | 아니오 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
