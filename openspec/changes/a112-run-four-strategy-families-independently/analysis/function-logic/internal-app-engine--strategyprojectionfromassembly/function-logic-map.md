# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `4302edefe72942bd1f4f7f4aa51b3c03e26ef97c13c3d8d52a0b6be94a2eef9f`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `106:1`–`177:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시장 레코드는 편집 전과 같은 값이다(첫 범위만). 범위 전부는 조정자 자식 selected[] 가 싣는다 — 같은 handoff 목록 · 같은 술어.

## Branches and early returns

- Exact AST return nodes: `176:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 109:2 | KR · US 순회 — **(편집)** 머리에서 조정자 자식을 채움(R4: 승인 범위 전부) |
| B2 | if | 119:3 | worker 미승격 → 시장 실패(조정자는 이미 채워짐) |
| B3 | switch | 121:4 | 실패 사유 고르기 |
| B4 | case | 122:4 | 활성화 부재 |
| B5 | case | 124:4 | 근거 stale |
| B6 | case | 126:4 | 보호 미배선 |
| B7 | range | 139:3 | 주문 경로와 같은 handoff 목록 순회 |
| B8 | if | 140:4 | 조정자 순서의 첫 승인 · 유효 범위(시장 레코드 — 불변) |
| B9 | if | 145:3 | 승인 범위 없음 → EvidenceStale |
| B10 | if | 153:3 | 레인 근거 digest 부재 → 후보 근거 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `assembly.Schedule.ObservedAt.UTC` | 107:14 |
| `strategyprojection.DormantSnapshot` | 108:14 |
| `strategyCoordinatorProjection` | 112:34 |
| `assembly.proposals.forMarket` | 112:72 |
| `strategyprojection.Market` | 113:23 |
| `assembly.Schedule.For` | 114:15 |
| `assembly.Candidate.For` | 115:16 |
| `assembly.Proposal.For` | 116:15 |
| `assembly.Risk.For` | 117:11 |
| `assembly.Supervisor.Snapshot` | 118:17 |
| `strategyprojection.WithMarketFailure` | 129:15 |
| `dispatchHandoffs` | 139:27 |
| `assembly.proposals.forMarket` | 139:27 |
| `handoff.Single` | 140:27 |
| `scoped.ValidProposal` | 140:57 |
| `strategyprojection.WithMarketFailure` | 146:15 |
| `projectionDigest` | 152:58 |
| `projectionDigest` | 154:21 |
| `strconv.Itoa` | 156:44 |
| `string` | 157:28 |
| `string` | 158:59 |
| `projectionDigest` | 159:23 |

## State mutations and fallbacks

- 상태 변경 없음 — 조립 값에서 스냅숏을 만든다.

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다.
