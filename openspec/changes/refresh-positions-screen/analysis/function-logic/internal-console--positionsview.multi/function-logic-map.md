# Function Logic Map: `positionsView.Multi`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

기존 1행 함수 — 이 change의 수정 대상이 아니며 인접 삽입(AnyUnknown 등)의 diff hunk 교차로 evidence가 요구되었다. 본문은 변경 전과 byte 동일하다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Accounts` | []string | `positions()`가 masked ref로 채움 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `len(v.Accounts) > 1` | 기존 다계좌 렌더 케이스(portfolio_test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음(순수 함수·본문 무변경).

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)
