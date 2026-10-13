# Branch Test Map: `computeTradeOutcome`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:165` `!model.Configured()` | TestAnAnalyticsFailureDoesNotRollBackTheClose(Costs 없음) | no | no |
| B2 | `:180` positions Query 오류 | 없음(도달하려면 연결 결함) | no | no |
| B3 | `:183` 행 없음 | 없음 | no | no |
| B4 | `:187` Scan 오류 | 없음 | no | no |
| B5 | `:193` 결정 · 편입 둘 다 없음 | 없음 — base 커버리지 0(외부 포지션은 exit 상태가 없어 ApplyExitFill B3 에서 돌아감); a130 미리보기 시험이 첫 도달 | no | no |
| B6 | `:201` frozenExitFacts 실패 | 없음(exit 상태 없는 CLOSED 는 ApplyExitFill 이 B3 에서 먼저 돌아감) | no | no |
| B7 | `:206` 편입 경로 | adoption_outcome_test 전부 | no | no |
| B8 | `:208` else 결정 경로 | trade_outcomes_test 왕복 시험 | no | no |
| B9 | `:211` `!ok \|\| buy.quantity.Sign() <= 0` | TestAnUnpricedFillLeavesNoOutcome · TestAPartlyEngineSoldAdoptedPositionFreezesNothing | no | no |
| B10 | `:217` 매수 비용 추정 오류 | 없음 | no | no |
| B11 | `:221` 매도 비용 추정 오류 | 없음 | no | no |
| B12 | `:234` 분모 ≠ 0 | TestRealizedRIsNotThePriceR · TestAnAdoptedPositionTheEngineLiquidatedFreezesAnOutcome | no | no |

## a130 시나리오 (3판)

- 사유 반환 — `TestA130EachReachableRefusalSiteNamesItsOwnReason` · `TestA130TheReasonSetIsClosedAndTokenShaped` · 기존 거절 시험
- B12 — 분모 = 위험 × (편입 + 묶음) — `TestA130AQuantityRaisedAfterAdoptionFreezesWithPerLotIntake`(2배 아님) · `TestA130AnEngineEnteredRoundTripIsUnchanged`
