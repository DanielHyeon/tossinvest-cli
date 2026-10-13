# Branch Test Map: `computeTradeOutcome`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:171` `if !model.Configured() {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B2 | `:186` `if err != nil {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B3 | `:189` `if !rows.Next() {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B4 | `:193` `if err := rows.Scan(&account, &market, &symbol, &decisionID, &adoptionID,` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B5 | `:199` `if decisionID == "" && adoptionID == "" {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B6 | `:207` `if !ok {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B7 | `:215` `if adoptionID != "" {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B8 | `:217` `} else if buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisio` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B9 | `:217` `} else if buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisio` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B10 | `:220` `if reason != "" {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B11 | `:223` `if buy.quantity.Sign() <= 0 {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B12 | `:229` `if err != nil {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B13 | `:233` `if err != nil {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
| B14 | `:250` `if denominator.Sign() != 0 {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130AnEngineEnteredRoundTripIsUnchanged` · 기존 trade_outcomes_test | no | yes |
