# Branch Test Map: `computeTradeOutcome`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:171` `if !model.Configured() {` | `TestA130AMovedItemIsNotWritten` · `TestA130AnUnconfiguredModelIsReportedWithItsReason` · `TestA130BarPricedRowsAreDerivableFromTheLedger` · `TestA130EachReachableRefusalSiteNamesItsOwnReason` 외 5 | no | yes |
| B2 | `:186` `if err != nil {` | not-applicable: positions 질의 실패 — 건강한 원장에서 도달 불가; a130 은 반환을 사유 코드로 바꿨을 뿐(사유 집합은 TestA130TheRefusalReasonSetIsClosed 가 고정) | no | n/a |
| B3 | `:189` `if !rows.Next() {` | not-applicable: 포지션 행 없음 — 동결은 같은 트랜잭션이 막 닫은 포지션을, backfill 은 positions 에서 고른 모집단을 넘기므로 생산에서 없음; 사유 코드만 a130 | no | n/a |
| B4 | `:193` `if err := rows.Scan(&account, &market, &symbol, &decisionID, &adoptionID,` | not-applicable: positions 행 Scan 실패 — 열 형식은 스키마가 고정(STRICT); 건강한 원장에서 도달 불가, 사유 코드만 a130 | no | n/a |
| B5 | `:199` `if decisionID == "" && adoptionID == "" {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` | no | yes |
| B6 | `:207` `if !ok {` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` | no | yes |
| B7 | `:215` `if adoptionID != "" {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 28 | no | yes |
| B8 | `:217` `} else if buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisio` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestAnUnpricedFillLeavesNoOutcome` | no | yes |
| B9 | `:217` `} else if buy, sell, ok = roundTripLegs(ctx, r, account, market, symbol, decisio` | `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestAnUnpricedFillLeavesNoOutcome` | no | yes |
| B10 | `:220` `if reason != "" {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21 | no | yes |
| B11 | `:223` `if buy.quantity.Sign() <= 0 {` | not-applicable: 매수 다리 수량 ≤ 0 — 편입 수량은 AdoptPosition 이 0 을 거절하고(adoption.go) 엔진 매수 다리는 체결 수량의 합이라 건강한 원장에서 없음, 방어 갈래; 사유 코드만 a130 | no | n/a |
| B12 | `:229` `if err != nil {` | not-applicable: 비용 추정 실패 — 앞 분기(B5 · B6)가 미구성 모델을 먼저 거르고 구성된 모델은 양의 금액에 실패하지 않음; 사유 코드만 a130 | no | n/a |
| B13 | `:233` `if err != nil {` | not-applicable: B12 와 같음(매도 쪽) | no | n/a |
| B14 | `:250` `if denominator.Sign() != 0 {` | `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal` 외 19 | no | yes |
