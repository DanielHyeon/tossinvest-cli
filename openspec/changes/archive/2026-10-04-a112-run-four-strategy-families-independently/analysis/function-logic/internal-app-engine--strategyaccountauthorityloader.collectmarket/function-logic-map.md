# Function Logic Map: `strategyAccountAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `c29e90e2a1e9f04e531a1cc000caa4bcc6a97a0cd796a73845a97dd23e1de7ca`
- Signature: `strategyAccountAuthorityLoader.collectMarket(params=3, results=1)`
- Source range: `155:1`–`199:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 준비 판정은 편집 전과 같다 — 바뀐 것은 실패 원인의 운반뿐(판정 불변).

## Branches and early returns

- Exact AST return nodes: `157:3, 163:3, 166:3, 198:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 162:2 | 항목 없음 · 활성화 없는 시장의 항목 하나 아님/무효 |
| B2 | if | 165:2 | loader 구성 불완전 |
| B3 | if | 169:2 | 시장이 US 면 계좌 시장 US |
| B4 | range | 173:2 | 항목(범위)마다 적재 |
| B5 | if | 178:3 | 범위 키 · 제안 유효할 때만 적재 |
| B6 | switch | 183:4 | **(새)** 적재 결과 분기 |
| B7 | case | 184:4 | **(새)** 적재 실패 → 원인 운반 |
| B8 | if | 187:5 | **(새)** 실패 뒤 ctx 종료 → 원인 = ctx 오류(결함) |
| B9 | case | 190:4 | 적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비 |
| B10 | case | 192:4 | **(새)** 적재 성공인데 시장 · 매니페스트 불일치 → 결함 원인 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Verified` | 161:15 |
| `proposal.familyActivation` | 161:15 |
| `len` | 162:5 |
| `len` | 162:50 |
| `ValidProposal` | 162:81 |
| `proposal.entries.authority.Proposal` | 162:81 |
| `fail` | 163:10 |
| `len` | 165:27 |
| `fail` | 166:10 |
| `make` | 172:12 |
| `len` | 172:53 |
| `entry.authority.Proposal` | 174:13 |
| `strategyOwnerKeyOf` | 175:17 |
| `errors.New` | 177:11 |
| `result.ValidProposal` | 178:15 |
| `loader.load` | 180:22 |
| `ctx.Err` | 187:18 |
| `authority.Market` | 190:9 |
| `authority.ManifestDigest` | 190:48 |
| `errors.New` | 193:20 |
| `append` | 196:12 |
| `strategyAccountMarketFromScopes` | 198:9 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접. 새로 통과 · 새로 거절하는 입력 0(원인 분류는 1차 레그에서).
