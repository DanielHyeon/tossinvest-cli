# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `c9cdc65328bc621f04b109f103d4111d774a877c96565ec5d7cb1b3ee7cd60dc`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `111:1`–`182:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시장 레코드는 편집 전과 같은 값이다(첫 범위만). 범위 전부는 조정자 자식 selected[] 가 싣는다 — 같은 handoff 목록 · 같은 술어.

## Branches and early returns

- Exact AST return nodes: `181:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 114:2 | KR · US 순회 — **(편집)** 머리에서 조정자 자식을 채움(R4: 승인 범위 전부) |
| B2 | if | 124:3 | worker 미승격 → 시장 실패(조정자는 이미 채워짐) |
| B3 | switch | 126:4 | 실패 사유 고르기 |
| B4 | case | 127:4 | 활성화 부재 |
| B5 | case | 129:4 | 근거 stale |
| B6 | case | 131:4 | 보호 미배선 |
| B7 | range | 144:3 | 주문 경로와 같은 handoff 목록 순회 |
| B8 | if | 145:4 | 조정자 순서의 첫 승인 · 유효 범위(시장 레코드 — 불변) |
| B9 | if | 150:3 | 승인 범위 없음 → EvidenceStale |
| B10 | if | 158:3 | 레인 근거 digest 부재 → 후보 근거 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `assembly.Schedule.ObservedAt.UTC` | 112:14 |
| `strategyprojection.DormantSnapshot` | 113:14 |
| `strategyCoordinatorProjection` | 117:34 |
| `assembly.proposals.forMarket` | 117:72 |
| `strategyprojection.Market` | 118:23 |
| `assembly.Schedule.For` | 119:15 |
| `assembly.Candidate.For` | 120:16 |
| `assembly.Proposal.For` | 121:15 |
| `assembly.Risk.For` | 122:11 |
| `assembly.Supervisor.Snapshot` | 123:17 |
| `strategyprojection.WithMarketFailure` | 134:15 |
| `dispatchHandoffs` | 144:27 |
| `assembly.proposals.forMarket` | 144:27 |
| `handoff.Single` | 145:27 |
| `scoped.ValidProposal` | 145:57 |
| `strategyprojection.WithMarketFailure` | 151:15 |
| `projectionDigest` | 157:58 |
| `projectionDigest` | 159:21 |
| `strconv.Itoa` | 161:44 |
| `string` | 162:28 |
| `string` | 163:59 |
| `projectionDigest` | 164:23 |

## State mutations and fallbacks

- 상태 변경 없음 — 조립 값에서 스냅숏을 만든다.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)
