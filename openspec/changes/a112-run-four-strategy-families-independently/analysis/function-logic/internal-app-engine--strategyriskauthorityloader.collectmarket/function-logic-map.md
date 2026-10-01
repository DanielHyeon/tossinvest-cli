# Function Logic Map: `strategyRiskAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_risk_authority.go`
- Source SHA-256: `f7ad67b8ad584c1719e3d582af7796621b5ca520a5800302182e38424a425e79`
- Signature: `strategyRiskAuthorityLoader.collectMarket(params=4, results=1)`
- Source range: `187:1`–`231:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 준비 판정 불변 — 바뀐 것은 실패 원인의 운반뿐.

## Branches and early returns

- Exact AST return nodes: `191:3, 194:3, 197:3, 230:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 193:2 | 결과 권한 준비 안 됨 |
| B2 | if | 196:2 | 환율 준비 안 됨 |
| B3 | if | 200:2 | 시장이 US 면 버킷 시장 US |
| B4 | range | 206:2 | 범위마다 번들 하나 |
| B5 | if | 210:3 | 범위 키 정규화 성공 시에만 적재 |
| B6 | switch | 217:4 | **(새)** 적재 결과 분기 |
| B7 | case | 218:4 | **(새)** 적재 실패 → 적재기 원인 그대로 운반(범위 국소 신원 포함 여부는 1차 레그가 가름) |
| B8 | case | 221:4 | 적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비 |
| B9 | case | 224:4 | **(새)** 적재 성공인데 시장 · 계좌 · 시각 · 항목 수 불일치 → 결함 원인 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fail` | 194:10 |
| `fail` | 197:10 |
| `make` | 205:12 |
| `len` | 205:50 |
| `result.results` | 205:54 |
| `result.results` | 206:25 |
| `strategyOwnerKeyOf` | 207:17 |
| `errors.New` | 209:11 |
| `riskbucket.LoadProductionRiskSnapshotAuthority` | 211:19 |
| `bundle.Scope` | 216:13 |
| `string` | 221:9 |
| `string` | 221:33 |
| `scope.AsOf.Equal` | 222:5 |
| `len` | 222:44 |
| `bundle.Entries` | 222:48 |
| `errors.New` | 225:19 |
| `append` | 228:12 |
| `strategyRiskMarketFromScopes` | 230:9 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(위험 권한). 새로 통과 · 새로 거절하는 입력 0.
