# Function Logic Map: `Gateway.place`

- Source: `internal/execgw/gateway.go` (362-400)
- Qualified function: `Gateway.place`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` b4b020607839345a…)
- AST evidence: `ast.json` — AST branches 1, 반환 5, 호출 20
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

place 계획을 만들고 `plan.call` 클로저(394–397)를 정의한다. 클로저는 호출마다 같은 `keyed(key)` 로 PlaceIntent 를 만들므로 같은 key 면 같은 본문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| req.Intent | PlaceIntent | 호출자 | submit 이 검증 |
| strategy | nil 또는 능력 | PlaceClaimedStrategy | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 384:4, 392:3, 396:3, 398:53, 399:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 382:2` | if g.preflight != nil { | 아래 「State mutations」 | 소스 줄 382 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`g.trading.Place` 는 클로저 안에서만 불리고 클로저는 submit 이 부른다.

AST 호출 이름 전체: `PlaceWireBody`, `g.executeOpts`, `g.preflight.CheckPlace`, `g.submit`, `g.trading.Place`, `g.trading.PreviewPlace`, `keyed`, `strings.EqualFold`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`.

## State mutations and fallbacks

- 상태 없음 — 계획 구성만.

## Safety conclusion

- a165 는 이 함수를 바꾸지 않을 계획(재시도 판정은 submit 쪽). High-risk: yes(주문 경로).
