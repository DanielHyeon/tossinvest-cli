# Function Logic Map: `Gateway.Cancel`

- Source: `internal/execgw/gateway.go` (403-422)
- Qualified function: `Gateway.Cancel`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` b4b020607839345a…)
- AST evidence: `ast.json` — AST branches 0, 반환 2, 호출 13
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

cancel 계획. `plan.call` 은 `g.trading.Cancel` 이고 trading.Service.Cancel 이 GET 사전 확인(GetOrderAvailableActions) 뒤 POST 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| req.Intent.OrderID | 브로커 주문 id | 호출자 | submit 이 검증 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 419:3, 421:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `branchless happy path at 403:1` | 분기 없음 — 단일 경로 | 아래 | 반환 419, 421 | 패키지 스위트 |

## Calls and live bindings

`g.trading.Cancel` 1회(클로저).

AST 호출 이름 전체: `g.executeOpts`, `g.submit`, `g.trading.Cancel`, `g.trading.PreviewCancel`, `orderTypeFor`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`.

## State mutations and fallbacks

- 상태 없음 — 계획 구성만.

## Safety conclusion

- 편집 계획 없음(409 재시도 판정은 submit 쪽). High-risk: yes.
