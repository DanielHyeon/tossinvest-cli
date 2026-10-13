# Function Logic Map: `adoptedRoundTripLegs`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `488`–`519`, base `233d5d9c`)
- AST evidence: `ast.json`(편집 뒤 `source_sha256` ff51203717c0ed3c…) — branches 9 · returns 10 · defers 0 · go 0; 편집 전 `ast-base.json`(`source_sha256` 0bd43abf96107b99…, branches 5 · returns 6)
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

## a130 편집 계획 (4판)

결함 자리 = B5 의 등식이 **편입 시점 수량**과 비교. 4판(사용자 결정 2026-10-13 3차): 시그니처에 `prices IntakePrices` 를 더함. B1 뒤에 `adoptionEpoch`(새 파일 — 재편입 · 편입 뒤 감소 거절, 편입 뒤 증가 묶음 목록) → `priceLots`(묶음마다 그 분 1분봉 종가; prices nil → needs_intake_price, 봉 없음 → no_intake_candle). 매수 다리 = 편입 수량 × 편입 관측가 + Σ 묶음 × 종가. B4 뒤 `instanceSellIntegrity`(foreign · unattributed). B5 는 `sell_unpriced` 와 `sell.quantity.Cmp(buy.quantity)` → `sell_quantity_mismatch`. 편입 뒤 증가가 없으면 a130 이전과 같은 값.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 체결 트랜잭션은 편입 뒤 증가 · 감소 · 재편입 · 인스턴스 밖 매도를 거절(보수), 증가 포지션의 행은 backfill 만, 거절 기록 추가.

## 편집 뒤 (a130 4판)

- 편집 뒤 `520`–`565` · `source_sha256` ff51203717c0ed3c… — branches 9 · returns 10
  (편집 전 5 · 6, 편집 전 AST 는 `ast-base.json`). 위 분기 표의 좌표는 편집 전, 아래 BTM 은 편집 뒤 좌표.
