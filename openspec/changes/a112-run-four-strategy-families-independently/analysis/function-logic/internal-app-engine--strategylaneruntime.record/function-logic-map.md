# Function Logic Map: `record`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `333970fa5e15db9741cc10da836922c79763f4b1b87bf62ddbc1af2fba9c6462`
- Signature: `strategyLaneRuntime.record(params=4, results=0)`
- Source range: `344:1`–`363:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 칸 대입은 Lock 뒤 · 파도 증가 뒤 · defer Unlock — AST 핀.

## Branches and early returns

- Exact AST return nodes: `348:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 347:2 | 관측 없음 → 물결 아님(칸도 그대로) |
| B2 | if | 352:2 | 파도 맵 지연 생성 |
| B3 | if | 355:2 | 파도 증가(포화) |
| B4 | range | 359:2 | 관측 덮어쓰기 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 347:23 |
| `runtime.mu.Lock` | 350:2 |
| `runtime.mu.Unlock` | 351:8 |
| `make` | 353:19 |
| `uint64` | 355:30 |

## State mutations and fallbacks

- observed · waves · shadowCells.

## Safety conclusion

- 레인 관측 기록(편집 전과 같음) + 값 보관.
