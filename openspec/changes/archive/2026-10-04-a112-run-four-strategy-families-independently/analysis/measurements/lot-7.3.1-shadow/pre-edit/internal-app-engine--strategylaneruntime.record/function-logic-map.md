# Function Logic Map (편집 전): `record`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `strategyLaneRuntime.record(params=2, results=0)`
- Source range: `322:1`–`338:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5 ④ ⑤: 파도 증가와 같은 임계 구역에서 시장 칸 {wave, batch, activation} 을 한 번에 대입(보관 대입 한 자리).

## Branches and early returns

- Exact AST return nodes: `324:3`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 323:2 | `if runtime == nil // len(observations) == 0 {` |
| B2 | if | 328:2 | `if runtime.waves == nil {` |
| B3 | if | 331:2 | `if runtime.waves[market] < ^uint64(0) {` |
| B4 | range | 334:2 | `for _, observation := range observations {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 323:23 |
| `runtime.mu.Lock` | 326:2 |
| `runtime.mu.Unlock` | 327:8 |
| `make` | 329:19 |
| `uint64` | 331:30 |

## Safety conclusion

- 레인 관측 기록 — 파도 증가 · 관측 덮어쓰기 · 빈 관측 guard 무변경.
