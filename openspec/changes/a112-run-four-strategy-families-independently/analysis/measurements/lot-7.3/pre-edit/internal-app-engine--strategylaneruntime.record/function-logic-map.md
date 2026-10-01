# Function Logic Map (편집 전): `record`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Signature: `strategyLaneRuntime.record(params=1, results=0)`
- Source range: `277:1`–`286:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: 시장 인자를 받아 시장별 물결 번호를 올리고 이번 관측들에 찍는다(Q1=(B), 0 = 미관측).

## Branches and early returns

- Exact AST return nodes: `279:3`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 278:2 | `if runtime == nil // len(observations) == 0 {` |
| B2 | range | 283:2 | `for _, observation := range observations {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 278:23 |
| `runtime.mu.Lock` | 281:2 |
| `runtime.mu.Unlock` | 282:8 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
