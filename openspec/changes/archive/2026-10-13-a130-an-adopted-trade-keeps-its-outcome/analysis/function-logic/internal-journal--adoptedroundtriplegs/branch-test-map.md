# Branch Test Map: `adoptedRoundTripLegs`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:526` `if !ok {` | not-applicable: 편입 기록 읽기 실패 — 편입 포지션은 adoption_id 가 가리키는 행이 FK 로 있음; 건강한 원장에서 도달 불가(design D5) | no | n/a |
| B2 | `:530` `if reason != "" {` | `TestA130APostAdoptionDisposalBlocksASweptFlatten` · `TestA130AReadoptedPositionIsRefused` · `TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal` · `TestA130TheSameSecondBoundaryFailsClosed` 외 2 | no | yes |
| B3 | `:534` `if reason != "" {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130AMissingOrWrongMinuteBarIsRefused` 외 8 | no | yes |
| B4 | `:541` `if !ok {` | not-applicable: engineSellOrders 읽기 실패 — 건강한 원장에서 도달 불가(design D5, TestA130EachReachableRefusalSiteNamesItsOwnReason 주석) | no | n/a |
| B5 | `:544` `if len(orders) == 0 {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` | no | yes |
| B6 | `:549` `if !sumSellFills(ctx, r, orders, account, market, symbol, &sell) {` | not-applicable: sumSellFills 읽기 실패 — 건강한 원장에서 도달 불가(design D5) | no | n/a |
| B7 | `:552` `if !sell.priced {` | `TestA130AnUnpricedEngineSellIsReportedAsSuch` | no | yes |
| B8 | `:555` `if reason := instanceSellIntegrity(ctx, r, positionID, account, market, symbol, ` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` · `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell` 외 1 | no | yes |
| B9 | `:558` `if sell.quantity.Cmp(buy.quantity) != 0 {` | `TestA130TheSameSecondBoundaryFailsClosed` | no | yes |
