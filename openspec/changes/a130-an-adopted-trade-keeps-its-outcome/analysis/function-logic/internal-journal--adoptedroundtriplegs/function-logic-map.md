# Function Logic Map: `adoptedRoundTripLegs`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `488`–`519`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 0bd43abf96107b99…) — branches 5 · returns 6 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 편입 포지션의 합성 매수 다리(편입 수량 × 관측가)와 귀속된 엔진 매도 다리를 만든다. 매도 수량이 편입 수량과 같아야 성공.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `computeTradeOutcome` B7 1.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` · `account` · `market` · `symbol` | computeTradeOutcome 이 읽은 값 | positions | — |
| 편입 기록 | quantity > 0 · observed_price > 0 | position_adoptions | adoptedBasis false → B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:494` adoptedBasis 실패 | 없음 | false `:495` | 없음 |
| B2 | `:501` engineSellOrders 실패 | 없음 | false `:502` | 없음 |
| B3 | `:504` 엔진 매도 주문 0 | 없음 | false `:507` | TestAnAdoptedPositionWithNoEngineSellFreezesNothing |
| B4 | `:509` sumSellFills 실패 | 없음 | false `:510` | 없음 |
| B5 | `:512` `!sell.priced \|\| sell.quantity.Cmp(adopted) != 0` | 없음 | false `:516` | TestAPartlyEngineSoldAdoptedPositionFreezesNothing(수량 부족); 미관측 매도 시험 없음 |
| (fallthrough) | 매도 수량 = 편입 수량 | — | true `:518` | TestAnAdoptedPositionTheEngineLiquidatedFreezesAnOutcome · TestAdoptedOutcomeSumsEveryEngineSell · TestAdoptedOutcomeIncludesFlattenSells |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `adoptedBasis` | 편입 수량 · 관측가 | false → B1 | AST |
| `engineSellOrders` | 귀속 주문(명시 참조) | false → B2 | AST |
| `sumSellFills` | 매도 수량 · 대금 | false → B4 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 없음. buy.quantity = adopted, buy.notional = adopted × observed.

## a130 편집 계획 (3판)

결함 자리 = B5 의 등식이 **편입 시점 수량**과 비교. 3판(사용자 결정 2026-10-13 2차): B1 뒤에 `adoptionEpoch`(새 함수, 새 파일 — 재편입 · 편입 뒤 감소 거절, 편입 뒤 증가 묶음마다 조정 뒤 383초 안 첫 엔진 관측가)를 부르고, 매수 다리 = 편입 수량 × 편입 관측가 + Σ 묶음 × 관측가, B5 는 `sell_unpriced` 와 `sell.quantity.Cmp(buy.quantity) != 0` → `sell_quantity_mismatch` 로 가름. 반환형 bool → 사유. 편입 뒤 증가가 없으면 a130 이전과 같은 값.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 편입 뒤 증가 포지션은 묶음별 편입가로 새로 동결, 감소 · 재편입 · 관측 없음은 거절, 거절 기록 추가.
