# Branch Test Map: `adoptedRoundTripLegs`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:494` adoptedBasis 실패 | 없음 | no | no |
| B2 | `:501` engineSellOrders 실패 | 없음 | no | no |
| B3 | `:504` 엔진 매도 주문 0 | TestAnAdoptedPositionWithNoEngineSellFreezesNothing | no | no |
| B4 | `:509` sumSellFills 실패 | 없음 | no | no |
| B5 | `:512` `!sell.priced \|\| sell.quantity.Cmp(adopted) != 0` | TestAPartlyEngineSoldAdoptedPositionFreezesNothing(수량 부족); 미관측 매도 시험 없음 | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1 — 편입 기록 읽기 실패 → adoption_unreadable — 없음 — 무편집 의미
- (신규) — 편입 뒤 조정 읽기 실패 / 평균가 없음 — a130 adjustment_unpriced 시험
- B2/B4 — 읽기 실패 → engine_sell_unreadable — 없음
- B3 — 엔진 매도 0 → no_engine_sell — TestAnAdoptedPositionWithNoEngineSellFreezesNothing + a130 사유 단언
- B5 — 교차 → sell_quantity_mismatch / 미관측 → sell_unpriced — TestAPartlyEngineSoldAdoptedPositionFreezesNothing · a130 교차 시험(편입 뒤 증가 + 일부 매도)
- fallthrough — Q_sell = Q_buy — a130 편입 뒤 증가 시험 · 편입 전 조정 시험 · 기존 시험 불변
