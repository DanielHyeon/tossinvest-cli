# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `95474831b04d24c21d90d72aac7349fe0682d2cbee9beb02c3be6307ac9dc510`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `98:1`–`165:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 읽기 전용 화면 — 주문 · 토글 · 원장 쓰기 없음. 두 범위 시장은 첫 범위만 보인다(범위별 행은 review 잔여).

## Branches and early returns

- Exact AST return nodes: `164:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 101:2 | 편집 전과 같은 분기 |
| B2 | if | 108:3 | 편집 전과 같은 분기 |
| B3 | switch | 110:4 | 편집 전과 같은 분기 |
| B4 | case | 111:4 | 편집 전과 같은 분기 |
| B5 | case | 113:4 | 편집 전과 같은 분기 |
| B6 | case | 115:4 | 편집 전과 같은 분기 |
| B7 | range | 127:3 | **(새)** 주문 경로와 같은 handoff 목록 순회 |
| B8 | if | 128:4 | **(새)** 조정자 순서의 첫 승인 · 유효 범위를 보임 |
| B9 | if | 133:3 | 승인 범위 없음 → EvidenceStale(편집 전 B7 의 결과) |
| B10 | if | 141:3 | 레인 증거 다이제스트 부재(편집 전 B8) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `assembly.Schedule.ObservedAt.UTC` | 99:14 |
| `strategyprojection.DormantSnapshot` | 100:14 |
| `strategyprojection.Market` | 102:23 |
| `assembly.Schedule.For` | 103:15 |
| `assembly.Candidate.For` | 104:16 |
| `assembly.Proposal.For` | 105:15 |
| `assembly.Risk.For` | 106:11 |
| `assembly.Supervisor.Snapshot` | 107:17 |
| `strategyprojection.WithMarketFailure` | 118:15 |
| `dispatchHandoffs` | 127:27 |
| `assembly.proposals.forMarket` | 127:27 |
| `handoff.Single` | 128:27 |
| `scoped.ValidProposal` | 128:57 |
| `strategyprojection.WithMarketFailure` | 134:15 |
| `projectionDigest` | 140:58 |
| `projectionDigest` | 142:21 |
| `strconv.Itoa` | 144:44 |
| `string` | 145:28 |
| `string` | 146:59 |
| `projectionDigest` | 147:23 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 아님(화면). 활성화 없는 시장은 handoff 하나라 편집 전과 같은 값.
