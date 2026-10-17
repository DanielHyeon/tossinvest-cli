# Function Logic Map: `DormantSnapshot`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`
- Signature: `DormantSnapshot(params=1, results=1)`
- Source range: `178:1`–`184:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 기본 레인 표는 골든 서술자와 같고(시험이 골든 파일을 직접 읽음) 생산 레인 목록과 같다(strategyworker 시험).

## Branches and early returns

- Exact AST return nodes: `180:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `generatedAt.UTC` | 179:16 |
| `unknownMarket` | 181:13 |
| `unknownMarket` | 182:13 |
| `defaultLanes` | 183:12 |
| `defaultCoordinators` | 183:42 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
