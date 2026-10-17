# Function Logic Map (편집 전): `UnavailableSnapshot`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`
- Signature: `UnavailableSnapshot(params=1, results=1)`
- Source range: `183:1`–`189:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: DormantSnapshot 과 같은 기본 자식들을 싣는다.

## Branches and early returns

- Exact AST return nodes: `185:2`.

- 분기 없음(AST 열거 0).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `generatedAt.UTC` | 184:16 |
| `unknownMarket` | 186:13 |
| `unknownMarket` | 187:13 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
