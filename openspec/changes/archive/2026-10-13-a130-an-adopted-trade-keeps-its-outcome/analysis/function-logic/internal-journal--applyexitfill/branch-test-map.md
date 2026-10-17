# Branch Test Map: `ApplyExitFill`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:973` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B2 | `:976` `if !found {` | not-applicable: 무편집 갈래이고 이 패키지 시험 전체가 도달하지 않음(측정) — a130 이전부터의 시험 공백, a130 범위 밖으로 기록 | no | n/a |
| B3 | `:983` `if errors.Is(err, ErrExitStateNotFound) {` | `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` · `TestA130AMovedItemIsNotWritten` 외 21 | no | yes |
| B4 | `:988` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B5 | `:992` `if strings.EqualFold(strings.TrimSpace(where.Side), "SELL") {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37 | no | yes |
| B6 | `:995` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B7 | `:999` `if cmp, err := riskcalc.CompareDecimal(next, state.TakenRatioTotal); err != nil ` | not-applicable: 저장된 taken_ratio_total 이 소수가 아닐 때만 — 쓰는 쪽이 늘 정규화 소수를 씀; 무편집 갈래 | no | n/a |
| B8 | `:1002` `} else if cmp > 0 {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37 | no | yes |
| B9 | `:1002` `} else if cmp > 0 {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37 | no | yes |
| B10 | `:1003` `if err := tx.MoveTakenRatioTotal(ctx, where.PositionID, next); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B11 | `:1009` `if fill.Terminal && state.Pending() && state.PendingIntentID != "" &&` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 25 | no | yes |
| B12 | `:1011` `if err := tx.ResolvePending(ctx, where.PositionID); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B13 | `:1014` `if err := appendExitEventFromHook(ctx, tx, exitEventRow{` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 25 | no | yes |
| B14 | `:1023` `if where.PositionState == PositionClosed && !state.Completed {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 36 | no | yes |
| B15 | `:1024` `if err := markExitCompletedTx(ctx, tx, where.PositionID); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B16 | `:1036` `if reason := freezeTradeOutcomeTx(ctx, tx, where.PositionID, tx.costs); reason !` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21 | no | yes |
