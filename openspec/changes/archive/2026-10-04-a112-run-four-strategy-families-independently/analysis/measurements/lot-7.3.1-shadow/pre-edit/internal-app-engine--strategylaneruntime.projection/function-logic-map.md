# Function Logic Map (편집 전): `projection`

- Source: `internal/app/engine/strategy_lane_projection.go`
- Source SHA-256: `04dcd7ed4366c18c5ac8b8c0cc490b944f5287dee4db481c35cfec07e6173d70`
- Signature: `strategyLaneRuntime.projection(params=0, results=1)`
- Source range: `31:1`–`43:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5.1: runtime.clk.Now() 로 shadowObservationUsable(파도 등식 ∧ 미만료 ∧ 나이 상한) 판정 후 레인 투영에 shadow 관측 전달.

## Branches and early returns

- Exact AST return nodes: `33:3`, `42:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 32:2 | `if runtime == nil {` |
| B2 | range | 38:2 | `for _, lane := range runtime.lanes {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.mu.RLock` | 35:2 |
| `runtime.mu.RUnlock` | 36:8 |
| `make` | 37:12 |
| `len` | 37:64 |
| `lane.Key` | 39:41 |
| `append` | 40:12 |
| `strategyLaneProjection` | 40:27 |

## Safety conclusion

- 읽기 전용 투영 — 레인 상태 · 관측을 바꾸지 않음(Offer · Fail · record 호출 0), 잠금 순서 런타임 → 레인 유지.
