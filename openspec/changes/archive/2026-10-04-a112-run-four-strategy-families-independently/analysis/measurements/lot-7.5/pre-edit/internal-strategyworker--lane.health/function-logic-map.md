# Function Logic Map (편집 전): `Health`

- Source: `internal/strategyworker/lane.go`
- Source SHA-256: `b6919f2ac3ce70c08631286b8c879bd3b3ab273228d21246b359fa5031e594c5`
- Signature: `Lane.Health(params=0, results=1)`
- Source range: `185:1`–`195:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.5 D2 — HEAD 69deeb48 사본에서 렌더(편집이 먼저 일어나 사본으로 편집 전을 잼)).

## Inputs and invariants

- 편집 계획: 판정 본문을 healthLocked 로 옮기고 Health 는 잠금 + 위임. Status 가 같은 잠금 안에서 같은 판정을 씀. 분기 2 → 0(본문 이동 — AST 열거 실측).

## Branches and early returns

- Exact AST return nodes: `189:3`, `192:3`, `194:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 188:2 | `if lane.latched {` |
| B2 | if | 191:2 | `if lane.consecutiveFailures > 0 {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.mu.Lock` | 186:2 |
| `lane.mu.Unlock` | 187:8 |

## Safety conclusion

- High-risk 아님 — 레인 건강 읽기(값 불변).
