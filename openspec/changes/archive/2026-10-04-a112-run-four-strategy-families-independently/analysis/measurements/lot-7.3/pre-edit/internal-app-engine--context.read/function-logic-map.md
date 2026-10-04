# Function Logic Map (편집 전): `Read`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `95474831b04d24c21d90d72aac7349fe0682d2cbee9beb02c3be6307ac9dc510`
- Signature: `Context.Read(params=1, results=2)`
- Source range: `24:1`–`66:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: 잠금 overlay 뒤에 생산 레인 런타임의 여덟 레인 관측을 읽기 전용으로 덧씌운다(lanes[8]). 기존 분기 불변.

## Branches and early returns

- Exact AST return nodes: `26:3`, `32:3`, `36:3`, `52:3`, `65:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 25:2 | `if c == nil // ctx == nil {` |
| B2 | if | 31:2 | `if store == nil {` |
| B3 | if | 35:2 | `if err != nil {` |
| B4 | if | 51:2 | `if supervisor == nil {` |
| B5 | range | 54:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B6 | if | 56:3 | `if !ok // !worker.Latched {` |
| B7 | if | 60:3 | `if current := snapshot.Markets[projectionMarket]; current.Status == strategyprojection.StatusCurrent {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 26:41 |
| `c.strategyProjectionMu.RLock` | 28:2 |
| `c.strategyProjectionMu.RUnlock` | 30:2 |
| `errors.New` | 32:41 |
| `store.Read` | 34:19 |
| `strategyprojection.WithRuntimeIdentity` | 49:13 |
| `strategyRuntimeConfigDigest` | 50:3 |
| `strategyRuntimeBuildDigest` | 50:34 |
| `supervisor.Snapshot` | 55:17 |
| `strategyprojection.Market` | 59:23 |
| `strategyprojection.WithMarketFailure` | 61:15 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
