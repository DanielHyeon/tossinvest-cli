# Function Logic Map: `parseOrderFacts`

- Source: `internal/execgw/indoubt.go` (591-623)
- Qualified function: `parseOrderFacts`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` d220e7df90d725bf…)
- AST evidence: `ast.json` — AST branches 6, 반환 5, 호출 17
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

브로커 주문 JSON → 매처 입력. quantity 결손은 `parseOptionalDecimal` 이 0 으로(B3), 그러면 `matcher.matches` B3 가 불일치로 판정 — a165 조회는 0 을 「모름 = 일치」로 다룬다(1.3 A8).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| raw | 주문 JSON | OrdersPageRaw | B1·B2 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 594:3, 597:3, 611:3, 614:3, 622:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 593:2` | if err := json.Unmarshal(raw, &payload); err != nil { | 아래 「State mutations」 | 소스 줄 593 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 596:2` | if strings.TrimSpace(payload.OrderID) == "" { | 아래 「State mutations」 | 소스 줄 596 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 610:2` | if facts.Quantity, err = parseOptionalDecimal(payload.Quantity); err != nil { | 아래 「State mutations」 | 소스 줄 610 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 613:2` | if facts.Price, err = parseOptionalDecimal(payload.Price); err != nil { | 아래 「State mutations」 | 소스 줄 613 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `if at 616:2` | if ts := strings.TrimSpace(payload.OrderedAt); ts != "" { | 아래 「State mutations」 | 소스 줄 616 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 617:3` | if parsed, perr := time.Parse(time.RFC3339, ts); perr == nil { | 아래 「State mutations」 | 소스 줄 617 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`json.Unmarshal`, `parseOptionalDecimal`, `time.Parse`.

AST 호출 이름 전체: `fmt.Errorf`, `json.Unmarshal`, `parseOptionalDecimal`, `parsed.UTC`, `strings.ToUpper`, `strings.TrimSpace`, `time.Parse`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
