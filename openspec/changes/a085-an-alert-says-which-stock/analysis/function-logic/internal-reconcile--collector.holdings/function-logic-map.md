# Function Logic Map: `Collector.holdings`

- Source: `internal/reconcile/snapshot.go` (lines 290–325)
- AST evidence: `ast.json` (`source_sha256: 9d93f80e6d9cb7d5d8d3819ac3aef812474ae94c04a3ffb1f4f2c168b828b674`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

이름을 나르는 배선. a085는 브로커 payload가 이미 담고 있던 `name`을 버리지 않고 다음 단계로 넘긴다. 비교·판정에는 쓰이지 않는 표시 전용 값이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (291) `if` — if raw, ok := c.Positions.(RawPositionsReader); ok | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (293) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (297) `range` — for _, h := range items | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (311) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (315) `range` — for _, p := range positions | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `raw.PositionsRaw` | (292) items, err := raw.PositionsRaw(ctx) | 호출부 계약 유지 | AST `calls` |
| `strings.ToUpper` | (299) Symbol:       strings.ToUpper(strings.TrimSpace(h.Symbol)), | 호출부 계약 유지 | AST `calls` |
| `strings.TrimSpace` | (299) Symbol:       strings.ToUpper(strings.TrimSpace(h.Symbol)), | 호출부 계약 유지 | AST `calls` |
| `canonicalDecimal` | (300) Quantity:     canonicalDecimal(h.Quantity), | 호출부 계약 유지 | AST `calls` |
| `strings.ToLower` | (302) Market:       strings.ToLower(strings.TrimSpace(h.Market)), | 호출부 계약 유지 | AST `calls` |
| `c.Positions.Positions` | (310) positions, err := c.Positions.Positions(ctx) | 호출부 계약 유지 | AST `calls` |
| `decimalString` | (318) Quantity:     decimalString(p.Quantity), | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
