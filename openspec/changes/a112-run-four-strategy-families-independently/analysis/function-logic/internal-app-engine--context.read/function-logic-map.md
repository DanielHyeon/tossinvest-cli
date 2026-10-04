# Function Logic Map: `Read`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `c9cdc65328bc621f04b109f103d4111d774a877c96565ec5d7cb1b3ee7cd60dc`
- Signature: `Context.Read(params=1, results=2)`
- Source range: `24:1`–`79:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- R1 시험 `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`(변이 S31).

## Branches and early returns

- Exact AST return nodes: `26:3, 32:3, 36:3, 60:3, 78:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 25:2 | Context · ctx nil |
| B2 | if | 31:2 | 투영 저장소 없음 |
| B3 | if | 35:2 | 저장소 읽기 오류 |
| B4 | if | 56:2 | 레인 런타임 있음 → 레인 덧씌움 |
| B5 | if | 59:2 | supervisor 없음 |
| B6 | range | 62:2 | 시장 순회 |
| B7 | if | 66:3 | **(새)** abandon 기록 시장 → SHADOW 지움 |
| B8 | if | 69:3 | worker 없음 · 잠기지 않음(편집 전 B7) |
| B9 | if | 73:3 | 현재 시장 레코드 → 실패 표시(편집 전 B8) |

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
| `c.strategyLanesMu.Lock` | 53:2 |
| `c.strategyLanesMu.Unlock` | 55:2 |
| `lanes.projection` | 57:20 |
| `supervisor.Snapshot` | 63:17 |
| `strategyLanesWithoutShadow` | 67:21 |
| `strategyprojection.Market` | 72:23 |
| `strategyprojection.WithMarketFailure` | 74:15 |

## State mutations and fallbacks

- 없음(읽기).

## Safety conclusion

- 읽기 전용 투영.
