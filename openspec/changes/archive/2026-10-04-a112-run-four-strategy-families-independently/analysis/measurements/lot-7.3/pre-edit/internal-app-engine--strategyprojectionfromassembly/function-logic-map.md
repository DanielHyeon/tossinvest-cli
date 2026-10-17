# Function Logic Map (편집 전): `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Source SHA-256: `95474831b04d24c21d90d72aac7349fe0682d2cbee9beb02c3be6307ac9dc510`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `98:1`–`165:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: 시장마다 조정자 투영(coordinators[2]: 준비 · 사유 · 수 · 중재 거절 · 제안 집합 digest · 승인 범위 전부 selected[] — R4)을 채운다. 기존 시장 레코드(첫 범위)는 불변.

## Branches and early returns

- Exact AST return nodes: `164:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 101:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B2 | if | 108:3 | `if !ok // !worker.Effective {` |
| B3 | switch | 110:4 | `switch {` |
| B4 | case | 111:4 | `case schedule.DesiredEnabled && !schedule.Ready:` |
| B5 | case | 113:4 | `case schedule.Ready && (!candidate.Ready // !proposal.Ready // !risk.Ready):` |
| B6 | case | 115:4 | `case schedule.Ready:` |
| B7 | range | 127:3 | `for _, handoff := range assembly.proposals.forMarket(market).dispatchHandoffs() {` |
| B8 | if | 128:4 | `if scoped, admitted := handoff.Single(); admitted && scoped.ValidProposal() {` |
| B9 | if | 133:3 | `if !handedOff {` |
| B10 | if | 141:3 | `if evidenceDigest == "" {` |

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

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
