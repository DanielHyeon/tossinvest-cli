# Function Logic Map (편집 전): `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`

- Source: `internal/app/engine/a112_coordinator_test.go`
- Source SHA-256: `157ac586f802f30b4089bdbf5f1bc075af45e1b3a78e7e554b07a15e3f1a5832`
- Signature: `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne(params=1, results=0)`
- Source range: `44:1`–`64:2`
- AST evidence: `ast.json` — 편집 **전**(a112 base 재고정 1e25b3a3 — 8.5 응답 로트(6f5b0df6) 편집 전(178cc196)).

## Inputs and invariants

- 편집 계획: 8.5 응답 로트 ⑦: 판정 활성화 carry 단언 추가(B10/B12 — 보이스 3 P2-1); collectOverflowing 은 configure 가변 인자(관문 아래 실행).

## Branches and early returns

- Exact AST return nodes: 없음.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 47:2 | `for index := range strategycoordinator.Capacity + 1 {` |
| B2 | if | 51:2 | `if pair.kr.snapshot.Reason != StrategyProposalQueueOverflow {` |
| B3 | if | 54:2 | `if pair.kr.snapshot.Ready // len(pair.kr.entries) != 0 {` |
| B4 | if | 57:2 | `if pair.kr.snapshot.QueueDropCount == 0 {` |
| B5 | if | 61:2 | `if pair.kr.snapshot.ArbitrationRefusal != "" {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 45:9 |
| `make` | 46:13 |
| `append` | 48:13 |
| `fmt.Sprintf` | 48:29 |
| `collectOverflowing` | 50:10 |
| `t.Fatalf` | 52:3 |
| `len` | 54:31 |
| `t.Fatalf` | 55:3 |
| `len` | 55:73 |
| `t.Fatalf` | 58:3 |
| `t.Fatalf` | 62:3 |

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
