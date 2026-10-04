# Function Logic Map (편집 전): `DormantSnapshot`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`
- Signature: `DormantSnapshot(params=1, results=1)`
- Source range: `175:1`–`181:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: 여덟 레인 OFF/OFF/UNOBSERVED · 두 조정자 미관측 기본값을 싣는다.

## Branches and early returns

- Exact AST return nodes: `177:2`.

- 분기 없음(AST 열거 0).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `generatedAt.UTC` | 176:16 |
| `unknownMarket` | 178:13 |
| `unknownMarket` | 179:13 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
