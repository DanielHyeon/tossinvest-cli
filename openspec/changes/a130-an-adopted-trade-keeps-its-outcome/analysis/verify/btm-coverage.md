# BTM 커버리지 측정 (a130 구현 리뷰 P2-2)

HEAD `2c2daae7` + 워킹트리(미커밋 6 경로) · `python3 analysis/harness/a130_btm_coverage.py` · covermode=count · 시험 하나씩 프로필. 「몸체 실행」 = 조건을 닫는 `{` 줄에서 시작하는 커버리지 블록 중 가장 오른쪽 것(몸체)의 횟수 > 0 (switch 는 그 줄을 품은 블록). 미도달은 원천을 읽은 사유가 있어야 rc 0.

## bindApplyHooks (`internal/app/engine/gateway.go`)

- B1 `:193` — `TestA130TheLoggedBindingCarriesTheSink`
- B2 `:196` — `TestA130TheLoggedBindingCarriesTheSink` · `TestBindingWiresBothApplyHooks` · `TestTheBoundExitApplierIsTheJournalsOwn` · `TestTheProjectionGuardAsksTheBindingAndNotTheData`

## NewContext (`internal/app/engine/engine.go`)

- B1 `:439` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B2 `:446` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B3 `:450` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B4 `:464` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B5 `:472` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B6 `:487` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B7 `:494` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B8 `:503` — not-applicable: bindApplyHooks 실패 — 새로 연 원장에 훅을 한 번 다는 자리라 「이미 달림」 거절이 생산에서 열리지 않음; a130 은 인자(싱크)만 더함, 거절 자체는 bindApplyHooks B1 · B2 가 잼
- B9 `:522` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B10 `:537` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B11 `:539` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B12 `:545` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B13 `:565` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B14 `:569` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B15 `:574` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)

## adoptedRoundTripLegs (`internal/journal/trade_outcomes.go`)

- B1 `:526` — not-applicable: 편입 기록 읽기 실패 — 편입 포지션은 adoption_id 가 가리키는 행이 FK 로 있음; 건강한 원장에서 도달 불가(design D5)
- B2 `:530` — `TestA130APostAdoptionDisposalBlocksASweptFlatten` · `TestA130AReadoptedPositionIsRefused` · `TestA130TheGrowThenShrinkRoundTripIsRefusedAsADisposal` · `TestA130TheSameSecondBoundaryFailsClosed` 외 2
- B3 `:534` — `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130AMissingOrWrongMinuteBarIsRefused` 외 8
- B4 `:541` — not-applicable: engineSellOrders 읽기 실패 — 건강한 원장에서 도달 불가(design D5, TestA130EachReachableRefusalSiteNamesItsOwnReason 주석)
- B5 `:544` — `TestA130EachReachableRefusalSiteNamesItsOwnReason`
- B6 `:549` — not-applicable: sumSellFills 읽기 실패 — 건강한 원장에서 도달 불가(design D5)
- B7 `:552` — `TestA130AnUnpricedEngineSellIsReportedAsSuch`
- B8 `:555` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` · `TestA130AnOverAttributedSellWithoutAnyAdjustmentIsAForeignSell` 외 1
- B9 `:558` — `TestA130TheSameSecondBoundaryFailsClosed`

## ApplyExitFill (`internal/journal/apply_hook.go`)

- B1 `:973` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B2 `:976` — not-applicable: 무편집 갈래이고 이 패키지 시험 전체가 도달하지 않음(측정) — a130 이전부터의 시험 공백, a130 범위 밖으로 기록
- B3 `:983` — `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal` · `TestA130ALocalSellOffsetBySweptFlattenIsNoLongerFrozen` · `TestA130AMovedItemIsNotWritten` 외 21
- B4 `:988` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B5 `:992` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37
- B6 `:995` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B7 `:999` — not-applicable: 저장된 taken_ratio_total 이 소수가 아닐 때만 — 쓰는 쪽이 늘 정규화 소수를 씀; 무편집 갈래
- B8 `:1002` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37
- B9 `:1002` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 37
- B10 `:1003` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B11 `:1009` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 25
- B12 `:1011` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B13 `:1014` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 25
- B14 `:1023` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 36
- B15 `:1024` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B16 `:1036` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21

## computeTradeOutcome (`internal/journal/trade_outcomes.go`)

- B1 `:171` — `TestA130AMovedItemIsNotWritten` · `TestA130AnUnconfiguredModelIsReportedWithItsReason` · `TestA130BarPricedRowsAreDerivableFromTheLedger` · `TestA130EachReachableRefusalSiteNamesItsOwnReason` 외 5
- B2 `:186` — not-applicable: positions 질의 실패 — 건강한 원장에서 도달 불가; a130 은 반환을 사유 코드로 바꿨을 뿐(사유 집합은 TestA130TheRefusalReasonSetIsClosed 가 고정)
- B3 `:189` — not-applicable: 포지션 행 없음 — 동결은 같은 트랜잭션이 막 닫은 포지션을, backfill 은 positions 에서 고른 모집단을 넘기므로 생산에서 없음; 사유 코드만 a130
- B4 `:193` — not-applicable: positions 행 Scan 실패 — 열 형식은 스키마가 고정(STRICT); 건강한 원장에서 도달 불가, 사유 코드만 a130
- B5 `:199` — `TestA130EachReachableRefusalSiteNamesItsOwnReason`
- B6 `:207` — `TestA130EachReachableRefusalSiteNamesItsOwnReason`
- B7 `:215` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 28
- B8 `:217` — `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestAnUnpricedFillLeavesNoOutcome`
- B9 `:217` — `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestAnUnpricedFillLeavesNoOutcome`
- B10 `:220` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21
- B11 `:223` — not-applicable: 매수 다리 수량 ≤ 0 — 편입 수량은 AdoptPosition 이 0 을 거절하고(adoption.go) 엔진 매수 다리는 체결 수량의 합이라 건강한 원장에서 없음, 방어 갈래; 사유 코드만 a130
- B12 `:229` — not-applicable: 비용 추정 실패 — 앞 분기(B5 · B6)가 미구성 모델을 먼저 거르고 구성된 모델은 양의 금액에 실패하지 않음; 사유 코드만 a130
- B13 `:233` — not-applicable: B12 와 같음(매도 쪽)
- B14 `:250` — `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` · `TestA130ALocalSellBeforeTheAdoptionIsNotThisTripsDisposal` 외 19

## freezeTradeOutcomeTx (`internal/journal/trade_outcomes.go`)

- B1 `:131` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21
- B2 `:144` — not-applicable: INSERT 실패 — 행 존재는 앞에서 확인했고 같은 트랜잭션이 쓰기 잠금을 쥠; 사유 코드만 a130

## BackfillTradeOutcome (`internal/journal/trade_outcomes.go`)

- B1 `:892` — `TestABackfillRecoversTheGapAndRefusesToRewriteIt`
- B2 `:894` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B3 `:894` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B4 `:899` — `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` · `TestAnAdoptedPositionWithNoEngineSellFreezesNothing`
- B5 `:903` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B6 `:913` — not-applicable: 앞의 TradeOutcomeOf 확인과 INSERT 사이에 다른 쓰기가 끼어야 열림(경합 전용); a130 무편집 갈래

## RecordFill (`internal/journal/fills.go`)

- B1 `:314` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B2 `:318` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B3 `:324` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B4 `:328` — not-applicable: 무편집 갈래이고 이 패키지 시험 전체가 도달하지 않음(측정) — a130 이전부터의 시험 공백, a130 범위 밖으로 기록
- B5 `:340` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B6 `:346` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B7 `:347` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B8 `:348` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B9 `:350` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B10 `:356` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B11 `:364` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B12 `:368` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B13 `:374` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B14 `:380` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B15 `:385` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B16 `:392` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B17 `:401` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B18 `:405` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B19 `:413` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B20 `:420` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B21 `:421` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B22 `:437` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B23 `:441` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B24 `:442` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B25 `:464` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B26 `:465` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B27 `:504` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B28 `:512` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B29 `:516` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B30 `:526` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B31 `:536` — `TestNonConfirmedFillDoesNotEnterTradeOutcome`
- B32 `:527` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B33 `:530` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B34 `:533` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B35 `:536` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B36 `:550` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B37 `:551` — `TestA130ARolledBackCloseReportsNothing`
- B38 `:556` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)

## runApplyHooks (`internal/journal/apply_hook.go`)

- B1 `:276` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B2 `:283` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B3 `:284` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B4 `:288` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B5 `:289` — 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)
- B6 `:293` — not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)
- B7 `:297` — `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40
- B8 `:298` — `TestA130ARolledBackCloseReportsNothing`

사유 없는 미도달 분기: 0
