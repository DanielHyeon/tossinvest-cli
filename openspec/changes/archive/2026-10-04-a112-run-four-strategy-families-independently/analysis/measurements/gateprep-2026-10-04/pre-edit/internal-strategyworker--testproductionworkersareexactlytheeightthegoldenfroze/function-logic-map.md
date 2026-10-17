# Function Logic Map (편집 전): `TestProductionWorkersAreExactlyTheEightTheGoldenFroze`

- Source: `internal/strategyworker/golden_contract_test.go`
- Source SHA-256: `3c298101cd35118c8c7fe242233bfaac674ea1e519812473fde82b5820f6b58f`
- Signature: `TestProductionWorkersAreExactlyTheEightTheGoldenFroze(params=1, results=0)`
- Source range: `79:1`–`120:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 7e124a73 (8.8.4 로트 A 직전)).

## Inputs and invariants

- 편집 계획: 8.8.4 로트 A: 공허한 desired/effective 절 제거(행동 단언은 태그 시험으로)

## Branches and early returns

- Exact AST return nodes: 없음.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 83:2 | `if golden.WorkerCount != len(golden.Descriptors) {` |
| B2 | if | 87:2 | `if len(workers) != golden.WorkerCount {` |
| B3 | range | 92:2 | `for index, want := range golden.Descriptors {` |
| B4 | if | 95:3 | `if string(key.Market) != want.Market // string(key.Family) != want.Family //` |
| B5 | if | 101:3 | `if string(got.Horizon()) != want.Horizon {` |
| B6 | if | 113:3 | `if string(got.Desired(none)) != want.Desired // string(got.Effective(none)) != want.Effective //` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `readGolden` | 80:12 |
| `ProductionWorkers` | 81:13 |
| `len` | 83:27 |
| `t.Fatalf` | 84:3 |
| `len` | 85:24 |
| `len` | 87:5 |
| `t.Fatalf` | 88:3 |
| `len` | 89:24 |
| `got.Key` | 94:10 |
| `string` | 95:6 |
| `string` | 95:43 |
| `t.Errorf` | 97:4 |
| `string` | 101:6 |
| `got.Horizon` | 101:13 |
| `t.Errorf` | 102:4 |
| `got.Horizon` | 102:85 |
| `string` | 113:6 |
| `got.Desired` | 113:13 |
| `string` | 113:51 |
| `got.Effective` | 113:58 |
| `string` | 114:4 |
| `got.Runtime` | 114:11 |
| `t.Errorf` | 115:4 |
| `got.Desired` | 117:5 |
| `got.Effective` | 117:24 |
| `got.Runtime` | 117:45 |

## Safety conclusion

- 시험 코드 — 생산 경로 없음
