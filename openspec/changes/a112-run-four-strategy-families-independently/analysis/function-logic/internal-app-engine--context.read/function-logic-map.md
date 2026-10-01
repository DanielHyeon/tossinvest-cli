# Function Logic Map: `Context.Read`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `4302edefe72942bd1f4f7f4aa51b3c03e26ef97c13c3d8d52a0b6be94a2eef9f`
- Signature: `Context.Read(params=1, results=2)`
- Source range: `24:1`–`74:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 레인 덧씌움은 `strategyLanesMu` 아래에서 포인터만 읽고 잠금 밖에서 projection 을 부른다(레인 접근자 · 런타임 읽기 잠금만).
- 읽기 전용 불변: Read 를 18 번 불러도 레인 상태 · 관측 기록 · 원장 잠금이 같다(`TestReadingTheLaneProjectionNeverChangesALane`, 변이 P01 · P02 CAUGHT).

## Branches and early returns

- Exact AST return nodes: `26:3, 32:3, 36:3, 60:3, 73:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 25:2 | nil receiver/context → 오류 |
| B2 | if | 31:2 | store 부재 → 오류 |
| B3 | if | 35:2 | store 읽기 오류 → 그대로 |
| B4 | if | 56:2 | **(새)** 레인 런타임 있음 → `lanes = runtime.projection()`(읽기만), 없으면 저장소의 미관측 기본값 |
| B5 | if | 59:2 | 감독자 부재 → 반환(편집 전 B4) |
| B6 | range | 62:2 | KR · US latch 검사(편집 전 B5) |
| B7 | if | 64:3 | latch 안 된 시장 건너뛰기(편집 전 B6) |
| B8 | if | 68:3 | CURRENT 시장만 덮기(편집 전 B7) |

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
| `strategyprojection.Market` | 67:23 |
| `strategyprojection.WithMarketFailure` | 69:15 |

## State mutations and fallbacks

- 상태 변경 없음 — 저장소 사본 위에 identity · 레인 · latch overlay 를 얹어 돌려준다.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
