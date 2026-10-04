# Function Logic Map: `Clone`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`
- Signature: `Clone(params=1, results=1)`
- Source range: `224:1`–`232:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 사본의 포인터 · 목록을 바꿔도 원본이 바뀌지 않는다.

## Branches and early returns

- Exact AST return nodes: `231:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 228:2 | 시장 두 개 복사(불변) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `cloneRuntimeIdentity` | 226:12 |
| `make` | 226:61 |
| `len` | 226:95 |
| `cloneLanes` | 227:10 |
| `cloneCoordinators` | 227:52 |
| `cloneMarket` | 229:25 |

## State mutations and fallbacks

- 상태 변경 없음 — 새 값을 만든다.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
