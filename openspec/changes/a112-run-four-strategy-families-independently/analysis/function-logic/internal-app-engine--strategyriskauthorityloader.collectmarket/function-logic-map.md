# Function Logic Map: `strategyRiskAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_risk_authority.go`
- Source SHA-256: `90192bbebb241d780d2ab843a8c0536eda30596aeb976adadfbdd45623274746`
- Signature: `strategyRiskAuthorityLoader.collectMarket(params=4, results=1)`
- Source range: `188:1`–`234:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 준비 판정 불변 — 바뀐 것은 실패 원인의 운반뿐.

## Branches and early returns

- Exact AST return nodes: `192:3, 195:3, 198:3, 233:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 194:2 | 결과 권한 준비 안 됨 |
| B2 | if | 197:2 | 환율 준비 안 됨 |
| B3 | if | 201:2 | 시장이 US 면 버킷 시장 US |
| B4 | range | 207:2 | 범위마다 번들 하나 |
| B5 | if | 211:3 | 범위 키 정규화 성공 시에만 적재 |
| B6 | switch | 220:4 | **(새)** 적재 결과 분기 |
| B7 | case | 221:4 | **(새)** 적재 실패 → 적재기 원인 그대로 운반(범위 국소 신원 포함 여부는 1차 레그가 가름) |
| B8 | case | 224:4 | 적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비 |
| B9 | case | 227:4 | **(새)** 적재 성공인데 시장 · 계좌 · 시각 · 항목 수 불일치 → 결함 원인 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fail` | 195:10 |
| `fail` | 198:10 |
| `make` | 206:12 |
| `len` | 206:50 |
| `result.results` | 206:54 |
| `result.results` | 207:25 |
| `strategyOwnerKeyOf` | 208:17 |
| `errors.New` | 210:11 |
| `riskbucket.LoadProductionRiskSnapshotAuthority` | 212:19 |
| `bundle.Scope` | 219:13 |
| `string` | 224:9 |
| `string` | 224:33 |
| `scope.AsOf.Equal` | 225:5 |
| `len` | 225:44 |
| `bundle.Entries` | 225:48 |
| `errors.New` | 228:19 |
| `append` | 231:12 |
| `strategyRiskMarketFromScopes` | 233:9 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(위험 권한). 새로 통과 · 새로 거절하는 입력 0.
