# Function Logic Map: `Service.Cancel`

- Source: `internal/trading/service.go` (210-222)
- Qualified function: `Service.Cancel`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 631b2f02437e5106…)
- AST evidence: `ast.json` — AST branches 3, 반환 4, 호출 5
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

취소. B3(218) `GetOrderAvailableActions` GET 사전 확인 뒤 POST. 재전송마다 GET 이 한 번 더 나간다(rate 예산).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| intent | CancelIntent | Gateway.Cancel | guard 실패 B1 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 212:3, 215:3, 219:3, 221:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 211:2` | if err := s.guard(ctx, ActionCancel, s.PreviewCancel(intent), opts); err != nil { | 아래 「State mutations」 | 소스 줄 211 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 214:2` | if s.broker == nil { | 아래 「State mutations」 | 소스 줄 214 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 218:2` | if _, err := s.broker.GetOrderAvailableActions(ctx, intent.OrderID); err != nil { | 아래 「State mutations」 | 소스 줄 218 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`s.guard`, `GetOrderAvailableActions`, `CancelPendingOrder`, `withLineage`.

AST 호출 이름 전체: `s.PreviewCancel`, `s.broker.CancelPendingOrder`, `s.broker.GetOrderAvailableActions`, `s.guard`, `s.withLineage`.

## State mutations and fallbacks

- lineage 기록(성공 시).

## Safety conclusion

- 편집 계획 없음.
