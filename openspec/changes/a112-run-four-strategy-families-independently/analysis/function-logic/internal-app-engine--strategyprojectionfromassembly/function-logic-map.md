# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Current source SHA-256: `045bad629e9087a86fdbb9c67cb6f90e4e9ef898f1a37fb2aefbbeee30aa3296`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `97:1`–`154:2`
- AST evidence: `ast.json`, regenerated because the **file** changed; AST 분기 8개.
- Risk scan: `risk-pattern-report.md`.

## 이 갱신이 왜 일어났는가 — 함수는 안 바뀌었다

a112 L5 가 같은 파일의 `Context.Read` 를 편집했다. 이 함수의 **본문은 한 글자도 바뀌지 않았고**,
파일 해시와 줄 번호만 움직였다(편집 전 81:1–138:2 → 지금 97:1–154:2). 분기 수는 그대로 8개다.

`ast.json` 의 `source_sha256` 은 파일 단위 해시라서, 같은 파일의 다른 함수를 고치면 이 번들도
stale 이 된다. 그래서 재생성하고 좌표를 다시 적는다 — 지도가 가리키는 줄이 실제 줄이어야 한다.

## Inputs and invariants

- Inputs/results are the exact AST signature above.
- 불변식: OFF 기본값, family/horizon 없는 owner key, 선행 조건이 없으면 노출을 늘리는 dispatch 0.

## Branches and early returns

- Exact AST return nodes: `153:2`.

| Branch | AST kind | Source location | Edited by this lot |
|---|---|---|---|
| B1 | range | 100:2 | 이 lot 은 이 함수를 편집하지 않았다 |
| B2 | if | 107:3 | 이 lot 은 이 함수를 편집하지 않았다 |
| B3 | switch | 109:4 | 이 lot 은 이 함수를 편집하지 않았다 |
| B4 | case | 110:4 | 이 lot 은 이 함수를 편집하지 않았다 |
| B5 | case | 112:4 | 이 lot 은 이 함수를 편집하지 않았다 |
| B6 | case | 114:4 | 이 lot 은 이 함수를 편집하지 않았다 |
| B7 | if | 121:3 | 이 lot 은 이 함수를 편집하지 않았다 |
| B8 | if | 130:3 | 이 lot 은 이 함수를 편집하지 않았다 |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| assembly.Schedule.ObservedAt.UTC | 98:14 |
| strategyprojection.DormantSnapshot | 99:14 |
| strategyprojection.Market | 101:23 |
| assembly.Schedule.For | 102:15 |
| assembly.Candidate.For | 103:16 |
| assembly.Proposal.For | 104:15 |
| assembly.Risk.For | 105:11 |
| assembly.Supervisor.Snapshot | 106:17 |
| strategyprojection.WithMarketFailure | 117:15 |
| assembly.proposals.forMarket | 120:16 |
| len | 121:6 |
| ValidProposal | 121:38 |
| authority.entries.authority.Proposal | 121:38 |
| strategyprojection.WithMarketFailure | 122:15 |
| authority.entries.authority.Proposal | 126:13 |
| projectionDigest | 129:58 |
| projectionDigest | 131:21 |
| strconv.Itoa | 133:44 |
| string | 134:28 |
| string | 135:59 |
| projectionDigest | 136:23 |

## State mutations and fallbacks

- AST 가 현재 base 의 assignment·call·branch·defer·return 전수 기록이다. 본문을 편집하는 lot 은
  바뀐 조건 의미와 RED/GREEN 증거로 이 지도를 갱신해야 한다.

## Safety conclusion

- 이 lot 은 이 함수를 편집하지 않았다. 좌표 갱신뿐이며 새 분기 주장도, 새 테스트 주장도 없다.
