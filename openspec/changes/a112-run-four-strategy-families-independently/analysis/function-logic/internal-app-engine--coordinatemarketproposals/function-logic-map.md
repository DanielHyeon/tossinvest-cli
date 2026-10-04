# Function Logic Map: `coordinateMarketProposals`

- Source: `internal/app/engine/strategy_market_coordinator.go`
- Source SHA-256: `590318d7267c1ffd5398c4c9878b9e70486633dc42033e19466e3474cf6da01b`
- Signature: `coordinateMarketProposals(params=6, results=3)`
- Source range: `57:1`–`130:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 조정 경로(admit · Submit · 계보 색인 · Arbitrate)의 식에 shadow 사용 0 — AST 핀.

## Branches and early returns

- Exact AST return nodes: `120:5, 129:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 66:2 | 경로(종목) 순회 |
| B2 | if | 68:3 | 레인 없는 종목 → 거절 계수 |
| B3 | range | 77:3 | 그 종목의 레인 제안 순회 — **편집: admit 앞에서 shadow.collect** |
| B4 | if | 96:4 | 관문이 멈춤 → gated 기록 |
| B5 | if | 101:4 | 관문 EMITTED → 레인 봉투 |
| B6 | if | 105:4 | 조정자 Submit 거절 |
| B7 | if | 109:4 | 계보 신원 충돌 → **부재 값**으로 반환(편집) |
| B8 | if | 124:3 | 범위 전부 관문에 빼앗김 → erasedScopes |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyRouterMarket` | 60:18 |
| `strategycoordinator.NewMarketCoordinator` | 61:17 |
| `make` | 62:55 |
| `batch.Len` | 62:103 |
| `batch.LanesFor` | 67:12 |
| `route.approved.Symbol` | 67:27 |
| `len` | 68:6 |
| `route.approved.Symbol` | 73:12 |
| `route.route.Request` | 73:57 |
| `lane.Proposal` | 78:14 |
| `lane.SnapshotDigest` | 82:21 |
| `shadow.collect` | 93:4 |
| `gate.admit` | 94:29 |
| `lane.SnapshotDigest` | 95:21 |
| `append` | 97:25 |
| `coordinator.Submit` | 104:17 |
| `coordinator.Drops` | 117:33 |
| `len` | 124:22 |
| `coordinator.Arbitrate` | 128:24 |

## State mutations and fallbacks

- 조정자 · arbitration · 지역 shadow 묶음만 쓴다(원장 · 브로커 0).

## Safety conclusion

- High-risk(조정 · 제안 권한 · 조립) 경로의 편집은 운반뿐이다 — 조정 · admit · Submit · Arbitrate · dispatch 의 입력 · 순서 · 반환은 편집 전과 같고(차등 dispatch 시험 · 변이 S01~S08), shadow 값은 authority 구조체에 들어가지 않는다(census ② `TestOnlyTheAllowedFunctionsEverTouchAShadowType`).
