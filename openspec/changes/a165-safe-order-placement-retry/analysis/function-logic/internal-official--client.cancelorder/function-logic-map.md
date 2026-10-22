# Function Logic Map: `Client.CancelOrder`

- Source: `internal/official/orders_write.go` (222-237)
- Qualified function: `Client.CancelOrder`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` d89e0bdbffdc2a41…)
- AST evidence: `ast.json` — AST branches 2, 반환 3, 호출 3
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

POST /api/v1/orders/{id}/cancel. 409 는 `postWithHeaders`→`send`→`classifyStatus` default 의 `*APIError{409, Body}` 로 올라온다(B2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| orderID | string | officialBroker | B1 계좌 헤더 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 226:3, 230:3, 232:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 225:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 225 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 229:2` | if err := c.postWithHeaders(ctx, path, struct{}{}, hdrs, &resp); err != nil { | 아래 「State mutations」 | 소스 줄 229 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`c.acctHeaders`, `c.postWithHeaders`.

AST 호출 이름 전체: `c.acctHeaders`, `c.postWithHeaders`, `url.PathEscape`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
