# Branch Test Map: `freezeTradeOutcomeTx`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:131` `!ok`(computeTradeOutcome 거절) | 거절 시험(TestAnAnalyticsFailureDoesNotRollBackTheClose · TestAnUnpricedFillLeavesNoOutcome · TestAPartlyEngineSoldAdoptedPositionFreezesNothing) | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1 — 거절 → 사유 반환 — 새 시험 a130(거절 사유) + 기존 거절 시험
- fallthrough — INSERT 성공 → "" / 실패 → insert_failed — 기존 TestTheOutcomeIsFrozenInTheClosingTransaction; insert_failed 는 변이로만(시험 픽스처로 INSERT 실패를 만들 길 없음 — UNIQUE 위반은 CLOSED 가 한 번뿐이라 도달 불가)
