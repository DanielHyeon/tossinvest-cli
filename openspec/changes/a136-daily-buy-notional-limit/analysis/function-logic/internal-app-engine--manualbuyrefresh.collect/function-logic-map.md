# Function Logic Map: `collect`

- Source: `internal/app/engine/manual_buy_refresh.go` (123-156, 편집 뒤)
- Qualified: `manualBuyRefresh.collect`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` db08e382da12e20f…, 편집 전 분기 수 8) — 현재 branches 8 · returns 7
- Risk scan: `risk-pattern-report.md`

**역할.** 수동 매수 refresher 의 한 번 읽기.

**a136 편집.** 844a042c: CLOSED 조회에 `Limit: 100`(문서 최대) — 분기 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 125) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B2 | `if err != nil {` (if 129) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B3 | `if err != nil {` (if 134) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B4 | `for page := 0; ; page++ {` (for 138) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B5 | `if page >= manualBuyClosedPageCap {` (if 139) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B6 | `if err != nil {` (if 143) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B7 | `if !closed.HasNext {` (if 147) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |
| B8 | `if closed.NextCursor == "" \|\| closed.NextCursor == cursor {` (if 150) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestA136AFailedReadIsPublishedAsUnknown` |

Exact AST return positions: 126 · 130 · 135 · 140 · 144 · 151 · 155

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 편집 내용 참조.

## Safety conclusion

- Safe edit boundary: 844a042c: CLOSED 조회에 `Limit: 100`(문서 최대) — 분기 무변경.
- High-risk impact: no(읽기 — 매수만 거부로 이어짐)
