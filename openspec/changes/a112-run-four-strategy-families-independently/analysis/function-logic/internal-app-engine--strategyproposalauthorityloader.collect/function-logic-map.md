# Function Logic Map: `collect`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `a349829128400c6ae6845aa49d271b318d582b0151532d2fcbd62515debfffbf`
- Signature: `strategyProposalAuthorityLoader.collect(params=4, results=2)`
- Source range: `258:1`–`297:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- recover 갈래의 부재 값 대입은 AST 핀(`TestTheRemainingCarryAndPublishSitesKeepTheirShape`).

## Branches and early returns

- Exact AST return nodes: `260:3, 296:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 259:2 | 입력 결함 → 실패 짝 + 부재 shadow 짝 |
| B2 | range | 268:2 | 시장 순회(KR · US goroutine) |
| B3 | if | 275:6 | recover — **편집: shadow 를 부재 값으로** |
| B4 | range | 288:2 | 결과 두 개 수신 |
| B5 | if | 290:3 | KR 결과 |
| B6 | else | 292:10 | US 결과 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.observedAt.IsZero` | 259:36 |
| `schedule.observedAt.Equal` | 259:69 |
| `schedule.observedAt.Equal` | 259:118 |
| `failedStrategyProposalPair` | 260:10 |
| `make` | 267:14 |
| `(unnamed)` | 270:6 |
| `(unnamed)` | 273:4 |
| `(unnamed)` | 274:11 |
| `recover` | 275:9 |
| `loader.collectMarket` | 281:13 |
| `schedule.forMarket` | 281:39 |
| `routes.forMarket` | 281:67 |
| `fx.forMarket` | 281:93 |

## State mutations and fallbacks

- 짝 두 개를 채운다(원장 0).

## Safety conclusion

- High-risk(조정 · 제안 권한 · 조립) 경로의 편집은 운반뿐이다 — 조정 · admit · Submit · Arbitrate · dispatch 의 입력 · 순서 · 반환은 편집 전과 같고(차등 dispatch 시험 · 변이 S01~S08), shadow 값은 authority 구조체에 들어가지 않는다(census ② `TestOnlyTheAllowedFunctionsEverTouchAShadowType`).

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
