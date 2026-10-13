# Branch Test Map: `adoptedRoundTripLegs`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:494` adoptedBasis 실패 | 없음 | no | no |
| B2 | `:501` engineSellOrders 실패 | 없음 | no | no |
| B3 | `:504` 엔진 매도 주문 0 | TestAnAdoptedPositionWithNoEngineSellFreezesNothing | no | no |
| B4 | `:509` sumSellFills 실패 | 없음 | no | no |
| B5 | `:512` `!sell.priced \|\| sell.quantity.Cmp(adopted) != 0` | TestAPartlyEngineSoldAdoptedPositionFreezesNothing(수량 부족); 미관측 매도 시험 없음 | no | no |

## a130 시나리오 (4판)

- B1 — adoption_unreadable — 무편집 의미
- (신규, B1 뒤) adoptionEpoch · priceLots — readopted · post_adoption_disposal · needs_intake_price · no_intake_candle · 묶음 종가 — `TestA130AReadoptedPositionIsRefused` · `TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal` · `TestA130APostAdoptionDisposalBlocksASweptFlatten` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AMissingOrWrongMinuteBarIsRefused` · `TestA130TheSameSecondBoundaryFailsClosed`
- B2/B4 — engine_sell_unreadable — 없음
- B3 — no_engine_sell — `TestA130EachReachableRefusalSiteNamesItsOwnReason/no_engine_sell`
- (신규, B4 뒤) instanceSellIntegrity — `TestA130AnUnattributedLocalSellIsRefused` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` · `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell`
- B5 — sell_unpriced / sell_quantity_mismatch — `TestA130AnUnpricedEngineSellIsReportedAsSuch` · same-second increase · TestAPartlyEngineSoldAdoptedPositionFreezesNothing
- 성공 — `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130APreAdoptionAdjustmentStillFreezes` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · 기존 adoption_outcome_test · trade_analytics_test
