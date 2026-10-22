# Function Logic Map: `Gateway.roundTripFor`

- Source: `internal/execgw/roundtrip.go` (71-78)
- Qualified function: `Gateway.roundTripFor`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` face1097e68e93a0…)
- AST evidence: `ast.json` — AST branches 1, 반환 3, 호출 1
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

ACK 뒤 정착 전 확인 훅. B1(72) place 가 아니거나 `g.orders` 가 nil 이면 nil. a165 A1 수리는 이 훅을 감싸 재전송 ACK 에 목록 대조를 더한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan | mutationPlan | submit | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 73:3, 75:2, 76:3

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 72:2` | if g.orders == nil \|\| plan.kind != journal.KindPlace { | 아래 「State mutations」 | 소스 줄 72 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`g.confirmCreatedOrder`.

AST 호출 이름 전체: `g.confirmCreatedOrder`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획: 감싸는 쪽(submit)에서 — 이 함수는 그대로.
