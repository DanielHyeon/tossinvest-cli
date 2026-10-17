# Function Logic Map: `strategyAccountAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `e6c12de7902b15de91de03a8da004f1ad167ce37d37580cf52766f4707ed0a4d`
- Signature: `strategyAccountAuthorityLoader.collectMarket(params=3, results=1)`
- Source range: `151:1`–`175:2`
- AST evidence: `ast.json` — **편집 전 현행**(a112 5.2.2.2 Pre-Edit, HEAD e8d56d49).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 입력: 한 시장의 제안 권한(조립이 중재한 항목 목록). 결과: 그 시장의 계좌 권한 하나(항목 하나의 종목으로 적재).
- **시장당 하나 가정:** B1 이 항목 수 1 을 요구하고 계좌 권한을 `entries[0]` 의 종목으로 적재한다(보이스 A #5 — 선택 범위와 대조되지 않음).

## Branches and early returns

- Exact AST return nodes: `153:3, 156:3, 159:3, 170:3, 172:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 155:2 | 제안 항목이 정확히 하나가 아니거나 그 제안이 유효하지 않음 → `StrategyAccountProposalNotReady`(시장당 하나 가정 — 5.2.2.2 가 범위별로 바꿈) |
| B2 | if | 158:2 | loader 구성 불완전(load · 키 · 설정 경로 · 계좌) → `StrategyAccountInternalFailure` |
| B3 | if | 163:2 | 시장이 US 면 계좌 시장 US |
| B4 | if | 169:2 | 계좌 권한 적재 실패 · 시장 · 매니페스트 불일치 → `StrategyAccountAuthorityUnavailable` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 155:5 |
| `ValidProposal` | 155:36 |
| `proposal.entries.authority.Proposal` | 155:36 |
| `fail` | 156:10 |
| `len` | 158:27 |
| `fail` | 159:10 |
| `proposal.entries.authority.Proposal` | 161:12 |
| `loader.load` | 166:20 |
| `authority.Market` | 169:19 |
| `authority.ManifestDigest` | 169:58 |
| `fail` | 170:10 |
| `authority.Generation` | 173:58 |
| `authority.QuoteCurrency` | 173:97 |
| `authority.ManifestDigest` | 174:19 |
| `authority.Identity` | 174:57 |

## State mutations and fallbacks

- 상태 쓰기 없음(계좌 권한 적재 = 읽기).

## Safety conclusion

- High-risk impact: yes(1차 레그 발급의 계좌 권한). 5.2.2.2 편집 전 기록.
