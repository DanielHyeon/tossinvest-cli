# Branch Test Map: `BackfillTradeOutcome`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:846` 이미 행 있음 | TestABackfillRecoversTheGapAndRefusesToRewriteIt | no | no |
| B2 | `:848` else | — | no | no |
| B3 | `:848` 조회 오류(NotFound 아님) | 없음 | no | no |
| B4 | `:853` compute 거절 | TestAnAdoptedPositionWithNoEngineSellFreezesNothing · TestAPartlyEngineSoldAdoptedPositionFreezesNothing | no | no |
| B5 | `:857` INSERT 오류 | 없음 | no | no |
| B6 | `:867` UNIQUE 위반 | 없음 | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1 — 재실행 → Exists — TestABackfillRecoversTheGapAndRefusesToRewriteIt
- B3 — 조회 오류 — 없음 — 무편집
- B4 — 거절 → 사유 코드가 오류 문장에 — a130 backfill 거절 시험 + 기존 둘
- B5/B6 — INSERT 오류 — 없음 — 무편집
