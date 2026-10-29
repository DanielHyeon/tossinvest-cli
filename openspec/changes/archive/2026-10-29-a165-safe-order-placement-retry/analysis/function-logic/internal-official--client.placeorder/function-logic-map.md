# Function Logic Map: `Client.PlaceOrder`

- Source: `internal/official/orders_write.go` (182-214)
- Qualified function: `Client.PlaceOrder`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` d89e0bdbffdc2a41…)
- AST evidence: `ast.json` — AST branches 4, 반환 5, 호출 5
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

POST /api/v1/orders. `buildOrderCreate`(B1) → `acctHeaders`(B2) → `postWithHeaders`(B3, 401 재전송은 그 안의 send) → echo 검사(B4). R5 시험의 대상.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| intent | PlaceIntent | officialBroker | B1·B2 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 185:3, 189:3, 193:3, 201:3, 205:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 184:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 184 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 188:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 188 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 192:2` | if err := c.postWithHeaders(ctx, "/api/v1/orders", body, hdrs, &resp); err != nil { | 아래 「State mutations」 | 소스 줄 192 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 200:2` | if intent.ClientOrderID != "" && resp.ClientOrderID != nil && *resp.ClientOrderID != intent.ClientOrderID { | 아래 「State mutations」 | 소스 줄 200 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`buildOrderCreate`, `c.acctHeaders`, `c.postWithHeaders`.

AST 호출 이름 전체: `buildOrderCreate`, `c.acctHeaders`, `c.postWithHeaders`, `fmt.Errorf`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
