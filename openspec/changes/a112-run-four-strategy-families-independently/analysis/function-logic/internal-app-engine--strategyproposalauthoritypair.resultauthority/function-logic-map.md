# Function Logic Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `dc5fbecc120d415e2c7607f6892741395d555b137e4fe6ad25fbf04c60011fa7`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- Source range: `181:1`–`209:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 결과 권한이 보는 범위 집합 = 주문 경로가 받는 handoff 집합(같은 함수).

## Branches and early returns

- Exact AST return nodes: `195:5, 200:4, 206:3, 208:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 192:3 | **(새)** handoff 목록 순회 |
| B2 | if | 194:4 | handoff 거절 또는 무효 제안 → 시장 결과 권한 준비 안 됨(편집 전 B1 — 목록의 일부만 넘기지 않음) |
| B3 | if | 199:3 | **(새)** 목록 없음 → 준비 안 됨 |
| B4 | if | 203:3 | **(새)** 범위 둘 이상 또는 활성화 시장 → 범위별 결과(`scoped`)를 싣는다 — 위험 적재기가 범위마다 번들을 만든다 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `value.dispatchHandoffs` | 190:15 |
| `make` | 191:14 |
| `len` | 191:45 |
| `handoff.Single` | 193:25 |
| `result.ValidProposal` | 194:22 |
| `append` | 197:14 |
| `len` | 199:6 |
| `len` | 203:6 |
| `Verified` | 203:26 |
| `value.familyActivation` | 203:26 |
| `convert` | 208:70 |
| `convert` | 208:110 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(위험 권한 입력). 활성화 없는 시장은 handoff 하나 — `scoped` 없음 → 위험 적재기가 편집 전과 같은 결과 하나를 본다.

a112 7.3 — 같은 파일 편집(관측 필드 · record 인자 / QueueDropCount 주석)으로 줄만 밀림

a112 8.8.4 로트 B: 같은 파일 collectMarket 편집(관문 계산 이동 · 주석)으로 줄만 밀림 — 본문 불변
