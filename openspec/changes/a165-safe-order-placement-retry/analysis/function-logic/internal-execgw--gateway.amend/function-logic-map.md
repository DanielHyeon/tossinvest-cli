# Function Logic Map: `Gateway.Amend`

- Source: `internal/execgw/gateway.go` (425-453)
- Qualified function: `Gateway.Amend`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` b4b020607839345a…)
- AST evidence: `ast.json` — AST branches 2, 반환 2, 호출 13
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

amend 계획. 생산 호출자는 CodeGraph 에서 0(ops `modifyHandler` 와 시험만).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| req.Intent | AmendIntent | 호출자 | submit 이 검증 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 450:3, 452:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 428:2` | if intent.Quantity != nil { | 아래 「State mutations」 | 소스 줄 428 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 432:2` | if intent.Price != nil { | 아래 「State mutations」 | 소스 줄 432 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`g.trading.Amend` 1회(클로저).

AST 호출 이름 전체: `g.executeOpts`, `g.submit`, `g.trading.Amend`, `g.trading.PreviewAmend`, `orderTypeFor`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`.

## State mutations and fallbacks

- 상태 없음.

## Safety conclusion

- 편집 계획 없음. High-risk: yes.
