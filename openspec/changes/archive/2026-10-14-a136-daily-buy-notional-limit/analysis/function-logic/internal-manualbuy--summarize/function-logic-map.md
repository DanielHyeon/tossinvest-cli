# Function Logic Map: `Summarize`

- Source: `internal/manualbuy/summary.go` (51-111, 편집 뒤)
- Qualified: `Summarize`
- AST evidence: `ast.json` (revision: base, 새 base 6cd36e6d) — 본문 무편집, 같은 파일의 이웃 편집(뒤에 덧붙인 시험 · 위의 문서 주석)으로 검사기가 교체된 base 함수로 분류해 편집 전 AST 를 요구함.
- Risk scan: `risk-pattern-report.md`

**역할.** 수동 매수 통화별 합산.

**a136 편집.** 844a042c: 문서 주석만(세부 상태 열거 · orderAmount 규칙) — 분기 16 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for i, raw := range raws {` (range 54) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B2 | `if err := json.Unmarshal(raw, &o); err != nil {` (if 56) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B3 | `if !strings.EqualFold(strings.TrimSpace(o.Side), "BUY") {` (if 59) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B4 | `if strings.TrimSpace(id) == "" {` (if 63) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B5 | `if engineOrders[id] {` (if 66) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B6 | `if err != nil {` (if 70) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B7 | `if err != nil {` (if 74) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B8 | `if label != day {` (if 77) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B9 | `if currency != "KRW" && currency != "USD" {` (if 81) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B10 | `if err != nil {` (if 85) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B11 | `if prev, seen := perOrder[id]; seen {` (if 88) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B12 | `if currencyOf[id] != currency {` (if 89) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B13 | `if prev.Cmp(amount) >= 0 {` (if 92) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B14 | `for id, amount := range perOrder {` (range 99) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B15 | `if totals[c] == nil {` (if 101) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |
| B16 | `for c, total := range totals {` (range 107) | — | 분기 본문 참조 | `TestA136ManualBuysAreSummedPerCurrency` · `TestA136ADuplicatedOrderCountsOnceAtTheLargerAmount` |

Exact AST return positions: 57 · 64 · 71 · 75 · 82 · 86 · 90 · 110

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 편집 내용 참조.

## Safety conclusion

- Safe edit boundary: 844a042c: 문서 주석만(세부 상태 열거 · orderAmount 규칙) — 분기 16 무변경.
- High-risk impact: yes(합계 입력)
