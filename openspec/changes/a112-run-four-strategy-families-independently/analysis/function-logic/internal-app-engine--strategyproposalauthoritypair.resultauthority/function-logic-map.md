# Function Logic Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `a356e5ead7d719e7b791423645a86b2a2f8b2eed26066127eafb1928ec411288`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- Source range: `183:1`–`211:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 결과 권한이 보는 범위 집합 = 주문 경로가 받는 handoff 집합(같은 함수).

## Branches and early returns

- Exact AST return nodes: `197:5, 202:4, 208:3, 210:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 194:3 | **(새)** handoff 목록 순회 |
| B2 | if | 196:4 | handoff 거절 또는 무효 제안 → 시장 결과 권한 준비 안 됨(편집 전 B1 — 목록의 일부만 넘기지 않음) |
| B3 | if | 201:3 | **(새)** 목록 없음 → 준비 안 됨 |
| B4 | if | 205:3 | **(새)** 범위 둘 이상 또는 활성화 시장 → 범위별 결과(`scoped`)를 싣는다 — 위험 적재기가 범위마다 번들을 만든다 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `value.dispatchHandoffs` | 192:15 |
| `make` | 193:14 |
| `len` | 193:45 |
| `handoff.Single` | 195:25 |
| `result.ValidProposal` | 196:22 |
| `append` | 199:14 |
| `len` | 201:6 |
| `len` | 205:6 |
| `Verified` | 205:26 |
| `value.familyActivation` | 205:26 |
| `convert` | 210:70 |
| `convert` | 210:110 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(위험 권한 입력). 활성화 없는 시장은 handoff 하나 — `scoped` 없음 → 위험 적재기가 편집 전과 같은 결과 하나를 본다.
