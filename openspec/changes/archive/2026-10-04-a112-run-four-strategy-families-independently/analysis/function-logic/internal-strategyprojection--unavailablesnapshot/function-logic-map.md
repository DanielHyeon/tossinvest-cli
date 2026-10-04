# Function Logic Map: `UnavailableSnapshot`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`
- Signature: `UnavailableSnapshot(params=1, results=1)`
- Source range: `186:1`–`192:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 미관측 자식은 사실을 싣지 않는다 — Validate B6 · B7 이 강제.

## Branches and early returns

- Exact AST return nodes: `188:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `generatedAt.UTC` | 187:16 |
| `unknownMarket` | 189:13 |
| `unknownMarket` | 190:13 |
| `defaultLanes` | 191:12 |
| `defaultCoordinators` | 191:42 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
