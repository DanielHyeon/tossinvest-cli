# Function Logic Map: `matcher.matches`

- Source: `internal/execgw/indoubt.go` (687-706)
- Qualified function: `matcher.matches`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` d220e7df90d725bf…)
- AST evidence: `ast.json` — AST branches 6, 반환 6, 호출 4
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

지문 비교. 모르는 값은 일치 쪽(B1·B2·B4·B5 의 빈 값 통과) — 단 수량(B3)은 결손 0 이 불일치가 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| o | orderFacts | parseOrderFacts | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 689:3, 692:3, 695:3, 698:3, 702:4, 705:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 688:2` | if m.symbol != "" && o.Symbol != "" && o.Symbol != m.symbol { | 아래 「State mutations」 | 소스 줄 688 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 691:2` | if m.side != "" && o.Side != "" && o.Side != m.side { | 아래 「State mutations」 | 소스 줄 691 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 694:2` | if !qtyClose(o.Quantity, m.quantity) { | 아래 「State mutations」 | 소스 줄 694 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `if at 697:2` | if m.price > 0 && o.Price > 0 && !qtyClose(o.Price, m.price) { | 아래 「State mutations」 | 소스 줄 697 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 700:2` | if m.useWindow && o.HasOrderedAt { | 아래 「State mutations」 | 소스 줄 700 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 701:3` | if o.OrderedAt.Before(m.windowStart) \|\| o.OrderedAt.After(m.windowEnd) { | 아래 「State mutations」 | 소스 줄 701 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`qtyClose`.

AST 호출 이름 전체: `o.OrderedAt.After`, `o.OrderedAt.Before`, `qtyClose`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
