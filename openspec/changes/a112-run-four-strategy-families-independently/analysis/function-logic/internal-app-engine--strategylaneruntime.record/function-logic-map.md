# Function Logic Map: `strategyLaneRuntime.record`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`
- Signature: `strategyLaneRuntime.record(params=2, results=0)`
- Source range: `293:1`–`309:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 증가와 찍기가 같은 쓰기 잠금 안이다 — 두 물결이 한 번호를 나눠 갖지 않는다.
- 포화 상한(^uint64(0))에서 멈춘다 — 넘쳐 0 이 되면 「미관측」으로 읽힌다.

## Branches and early returns

- Exact AST return nodes: `295:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 294:2 | nil 런타임 또는 관측 0 → 물결이 아님(번호 안 올림) |
| B2 | if | 299:2 | **(새)** 물결 맵 없음 → 생성(첫 물결) |
| B3 | if | 302:2 | **(새)** 포화 상한 아래면 +1 |
| B4 | range | 305:2 | 관측마다 물결 번호를 찍어 덮어쓰기 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 294:23 |
| `runtime.mu.Lock` | 297:2 |
| `runtime.mu.Unlock` | 298:8 |
| `make` | 300:19 |
| `uint64` | 302:30 |

## State mutations and fallbacks

- 런타임의 관측 맵 · 물결 맵(프로세스 메모리)만 쓴다.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
