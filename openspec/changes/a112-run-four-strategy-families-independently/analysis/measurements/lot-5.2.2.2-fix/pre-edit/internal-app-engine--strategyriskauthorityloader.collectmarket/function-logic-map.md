# Function Logic Map: `strategyRiskAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_risk_authority.go`
- Source SHA-256: `bd5589d0c7d35de8294af8646be3502615fc0509d43a27c19931fe3fcc093d10`
- Signature: `strategyRiskAuthorityLoader.collectMarket(params=4, results=1)`
- Source range: `186:1`–`223:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 번들은 모두 같은 서명 시장 매니페스트에서 온다(세대 · 정책은 시장 단위, 버킷 사용량 스냅숏은 적재 시점의 원장).

## Branches and early returns

- Exact AST return nodes: `190:3, 193:3, 196:3, 222:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 192:2 | 결과 권한 준비 안 됨 → LaneNotReady |
| B2 | if | 195:2 | 환율 준비 안 됨 → FXNotReady |
| B3 | if | 199:2 | 시장이 US 면 버킷 시장 US |
| B4 | range | 205:2 | **(새)** 결과 권한의 범위마다 번들 하나(`result.results()` — 활성화 없는 시장은 하나) |
| B5 | if | 208:3 | **(새)** 범위 키 정규화 성공 시에만 적재 — 실패면 그 범위 AuthorityUnavailable |
| B6 | if | 215:4 | 적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비(편집 전 B4 · B5 의 반대편) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fail` | 193:10 |
| `fail` | 196:10 |
| `make` | 204:12 |
| `len` | 204:50 |
| `result.results` | 204:54 |
| `result.results` | 205:25 |
| `strategyOwnerKeyOf` | 206:17 |
| `riskbucket.LoadProductionRiskSnapshotAuthority` | 209:19 |
| `bundle.Scope` | 214:13 |
| `string` | 215:21 |
| `string` | 215:45 |
| `scope.AsOf.Equal` | 216:5 |
| `len` | 216:44 |
| `bundle.Entries` | 216:48 |
| `append` | 220:12 |
| `strategyRiskMarketFromScopes` | 222:9 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(위험 권한). 한 범위 적재 실패는 그 범위만 준비 안 됨(J3). 오늘 생산에서는 스키마 핀 27 결함으로 전 범위가 준비 안 됨 — ROADMAP a112 이월.
