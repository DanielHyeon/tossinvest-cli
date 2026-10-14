# Function Logic Map: `valueOf`

- Source: `internal/manualbuy/summary.go` (122-171, 편집 뒤)
- Qualified: `valueOf`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` e4292f1ecee2c9c1…, 편집 전 분기 수 11) — 현재 branches 14 · returns 10
- Risk scan: `risk-pattern-report.md`

**역할.** 한 매수 주문의 금액.

**a136 편집.** 844a042c(구현 리뷰 P1-2 · P2-1): 그룹 라벨 "CLOSED" 비교를 OpenAPI 세부 상태 종결 집합(`terminalStatuses`)으로 바꿈 · 금액 기반 시장가 `orderAmount` 갈래 추가 · 수량 파싱을 미체결 갈래로 옮김 — 분기 11 → 14.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if o.Execution != nil {` (if 127) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B2 | `if filled, err = decimal("execution.filledQuantity", o.Execution.FilledQuantity, true); err != nil {` (if 128) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B3 | `switch {` (switch 131) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B4 | `case strings.TrimSpace(o.Execution.FilledAmount) != "":` (case 132) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B5 | `if value, err = decimal("execution.filledAmount", o.Execution.FilledAmount, false); err != nil {` (if 133) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B6 | `case filled.Sign() > 0:` (case 136) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B7 | `if err != nil {` (if 138) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B8 | `if terminalStatuses[status] {` (if 144) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B9 | `if strings.TrimSpace(o.OrderAmount) != "" {` (if 148) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B10 | `if err != nil {` (if 150) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B11 | `if amount.Cmp(value) > 0 {` (if 153) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B12 | `if err != nil {` (if 159) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B13 | `if remaining.Sign() > 0 {` (if 163) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |
| B14 | `if err != nil {` (if 165) | — | 분기 본문 참조 | `TestA136TerminalStatusesCountOnlyWhatFilled` · `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount` · `TestA136ManualBuysAreSummedPerCurrency` · `TestA136AnUnvaluableManualBuyFailsTheWholeSnapshot` |

Exact AST return positions: 129 · 134 · 139 · 145 · 151 · 154 · 156 · 160 · 166 · 170

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 편집 내용 참조.

## Safety conclusion

- Safe edit boundary: 844a042c(구현 리뷰 P1-2 · P2-1): 그룹 라벨 "CLOSED" 비교를 OpenAPI 세부 상태 종결 집합(`terminalStatuses`)으로 바꿈 · 금액 기반 시장가 `orderAmount` 갈래 추가 · 수량 파싱을 미체결 갈래로 옮김 — 분기 11 → 14.
- High-risk impact: yes(합계 입력)
