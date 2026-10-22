# Branch Test Map: `matcher.matches`

- Source: `internal/execgw/indoubt.go` (687-706)
- AST branches 6 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 688:2` | if m.symbol != "" && o.Symbol != "" && o.Symbol != m.symbol { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 691:2` | if m.side != "" && o.Side != "" && o.Side != m.side { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `if at 694:2` | if !qtyClose(o.Quantity, m.quantity) { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `if at 697:2` | if m.price > 0 && o.Price > 0 && !qtyClose(o.Price, m.price) { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `if at 700:2` | if m.useWindow && o.HasOrderedAt { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 701:3` | if o.OrderedAt.Before(m.windowStart) \|\| o.OrderedAt.After(m.windowEnd) { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
