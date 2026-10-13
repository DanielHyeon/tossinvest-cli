# Function Logic Map: `computeTradeOutcome`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `163`–`250`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 0bd43abf96107b99…) — branches 12 · returns 10 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 닫힌 포지션의 성과 행(손익 · 실현 R · 위험 · 수량 · 비용 · 보유 시간 · 도달 단계)을 계산한다. 쓰지 않는다.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `freezeTradeOutcomeTx` · `Journal.BackfillTradeOutcome`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r` | tx 또는 journalReader | 호출자 | Query 오류 → false |
| `positionID` | — | 호출자 | 없으면 B3 |
| `model` | costs.Model | 주입 | 미구성 → B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:165` `!model.Configured()` | 없음 | false `:169` | TestAnAnalyticsFailureDoesNotRollBackTheClose(Costs 없음) |
| B2 | `:180` positions Query 오류 | 없음 | false `:181` | 없음(도달하려면 연결 결함) |
| B3 | `:183` 행 없음 | rows.Close | false `:185` | 없음 |
| B4 | `:187` Scan 오류 | rows.Close | false `:190` | 없음 |
| B5 | `:193` 결정 · 편입 둘 다 없음 | 없음 | false `:197` | 없음 — base 커버리지 0(TestAnExternalPositionHasNoOutcome 은 exit 상태가 없어 ApplyExitFill B3 에서 먼저 돌아감) |
| B6 | `:201` frozenExitFacts 실패 | 없음 | false `:202` | 없음(exit 상태 없는 CLOSED 는 ApplyExitFill 이 B3 에서 먼저 돌아감) |
| B7 | `:206` 편입 경로 | — | → adoptedRoundTripLegs | adoption_outcome_test 전부 |
| B8 | `:208` else 결정 경로 | — | → roundTripLegs | trade_outcomes_test 왕복 시험 |
| B9 | `:211` `!ok \|\| buy.quantity.Sign() <= 0` | 없음 | false `:212` | TestAnUnpricedFillLeavesNoOutcome · TestAPartlyEngineSoldAdoptedPositionFreezesNothing |
| B10 | `:217` 매수 비용 추정 오류 | 없음 | false `:218` | 없음 |
| B11 | `:221` 매도 비용 추정 오류 | 없음 | false `:222` | 없음 |
| B12 | `:234` 분모 ≠ 0 | realizedR 계산 | 성공 `:238`(분모 0 이면 "0") | TestRealizedRIsNotThePriceR · TestAnAdoptedPositionTheEngineLiquidatedFreezesAnOutcome |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `frozenExitFacts` | initial_risk · 단계 | false → B6 | AST |
| `adoptedRoundTripLegs` / `roundTripLegs` | 다리 | false → B9 | AST |
| `model.EstimateTradeCost` ×2 | 비용 | 오류 → B10/B11 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 없음(읽기 · 계산만). 실현 R 분모 = `initial_risk × buy.quantity`(:232).

## a130 편집 계획 (4판)

반환형 (TradeOutcome, bool) → (TradeOutcome, TradeOutcomeRefusalReason), 시그니처에 `prices IntakePrices`(체결 트랜잭션 · BackfillTradeOutcome 은 nil). 거절 return 9개가 design D2 의 코드. B7 편입 경로는 adoptedRoundTripLegs 의 사유를 올림. 분모 `initial_risk × buy.quantity` 그대로 — 편입 경로에서 buy.quantity = 편입 + 묶음 합(= 매도 수량).

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 체결 트랜잭션은 편입 뒤 증가 · 감소 · 재편입 · 인스턴스 밖 매도를 거절(보수), 증가 포지션의 행은 backfill 만, 거절 기록 추가.
