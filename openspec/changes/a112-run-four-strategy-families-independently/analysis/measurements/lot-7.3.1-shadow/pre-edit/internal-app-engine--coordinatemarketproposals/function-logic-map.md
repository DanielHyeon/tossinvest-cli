# Function Logic Map (편집 전): `coordinateMarketProposals`

- Source: `internal/app/engine/strategy_market_coordinator.go`
- Source SHA-256: `e9a4bde458176c4679f254d71b56439b348110fb6dc1ab2338cd7f2cf376f728`
- Signature: `coordinateMarketProposals(params=6, results=2)`
- Source range: `57:1`–`125:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 v3.3 §4: 반환에 strategyShadowBatch 추가; 안쪽 루프 gate.admit 문장 앞에 수집 helper 호출 하나(주소 지정 가능한 지역 값 수신자); 계보 충돌 return 은 부재 값 리터럴, 정상 return 만 수집 묶음. 조정 경로(admit · Submit · 계보 색인 · Arbitrate) 식에 shadow 사용 0, 루프 순서 무변경.

## Branches and early returns

- Exact AST return nodes: `115:5`, `124:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 64:2 | `for _, route := range routes {` |
| B2 | if | 66:3 | `if len(lanes) == 0 {` |
| B3 | range | 75:3 | `for _, lane := range lanes {` |
| B4 | if | 93:4 | `if !ok {` |
| B5 | if | 98:4 | `if outcome == strategyworker.OutcomeEmitted {` |
| B6 | if | 102:4 | `if !admission.Admitted {` |
| B7 | if | 106:4 | `if _, exists := arbitration.byIdentity[result.Lineage.Identity]; exists {` |
| B8 | if | 119:3 | `if gatedInScope == len(lanes) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyRouterMarket` | 60:18 |
| `strategycoordinator.NewMarketCoordinator` | 61:17 |
| `make` | 62:55 |
| `batch.Len` | 62:103 |
| `batch.LanesFor` | 65:12 |
| `route.approved.Symbol` | 65:27 |
| `len` | 66:6 |
| `route.approved.Symbol` | 71:12 |
| `route.route.Request` | 71:57 |
| `lane.Proposal` | 76:14 |
| `lane.SnapshotDigest` | 80:21 |
| `gate.admit` | 91:29 |
| `lane.SnapshotDigest` | 92:21 |
| `append` | 94:25 |
| `coordinator.Submit` | 101:17 |
| `coordinator.Drops` | 114:33 |
| `len` | 119:22 |
| `coordinator.Arbitrate` | 123:24 |

## Safety conclusion

- High-risk(조정 경로) — 조정 · admit · Submit · Arbitrate 의 입력 · 순서 · 반환 무변경; 추가 문장은 panic 거리 없는 append 하나(shape 핀), 계보 충돌은 부재 값(부분 묶음 금지).
