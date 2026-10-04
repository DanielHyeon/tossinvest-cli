# Function Logic Map: `strategyAccountAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `d0d6281292dafcc979edce741a3a2bf98ed348f023267d8198d2436c71ec7291`
- Signature: `strategyAccountAuthorityLoader.collectMarket(params=3, results=1)`
- Source range: `154:1`–`188:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시장 칸(`authority` · `snapshot`)은 첫 준비된 범위의 것 — 범위 하나면 편집 전과 같은 값.
- 범위의 계좌 권한은 `forScope(key)` 로만 1차 레그에 건너간다(봉투 폴백 없음).

## Branches and early returns

- Exact AST return nodes: `156:3, 162:3, 165:3, 187:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 161:2 | 항목 없음, 또는 **활성화 없는 시장**에서 항목이 정확히 하나가 아니거나 무효 → `StrategyAccountProposalNotReady`(토글 OFF = upstream) |
| B2 | if | 164:2 | loader 구성 불완전 → `StrategyAccountInternalFailure` |
| B3 | if | 168:2 | 시장이 US 면 계좌 시장 US |
| B4 | range | 172:2 | **(새)** 항목(소유자 범위)마다 계좌 권한 적재 — 적재 종목은 `entries[0]` 이 아니라 **그 범위의 종목**(A#5) |
| B5 | if | 176:3 | **(새)** 범위 키 정규화 실패 또는 무효 제안 → 그 범위만 `ProposalNotReady` |
| B6 | if | 181:4 | 적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비(편집 전 B4 의 반대편); 실패는 그 범위만 `AuthorityUnavailable` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Verified` | 160:15 |
| `proposal.familyActivation` | 160:15 |
| `len` | 161:5 |
| `len` | 161:50 |
| `ValidProposal` | 161:81 |
| `proposal.entries.authority.Proposal` | 161:81 |
| `fail` | 162:10 |
| `len` | 164:27 |
| `fail` | 165:10 |
| `make` | 171:12 |
| `len` | 171:53 |
| `entry.authority.Proposal` | 173:13 |
| `strategyOwnerKeyOf` | 174:17 |
| `result.ValidProposal` | 176:15 |
| `loader.load` | 178:22 |
| `authority.Market` | 181:21 |
| `authority.ManifestDigest` | 181:60 |
| `append` | 185:12 |
| `strategyAccountMarketFromScopes` | 187:9 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(1차 레그의 계좌 권한 출처). 활성화 없는 시장의 새 통과 입력 0. 활성화 시장에서 한 범위의 적재 실패가 시장 전체를 닫지 않게 됐다(J3) — 그 범위는 1차 레그에서 타입 거절된다.
