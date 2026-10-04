# Function Logic Map: `TestProductionWorkersAreExactlyTheEightTheGoldenFroze (시험)`

- Source: `internal/strategyworker/golden_contract_test.go`
- Source SHA-256: `38997843a0a5276323cef0187e88945fc8368b1c0e245572007599a3718bd892`
- Signature: `TestProductionWorkersAreExactlyTheEightTheGoldenFroze(params=1, results=0)`
- Source range: `79:1`–`112:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 gate-prep).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 골든 서술자 순서 · 열쇠 · horizon · runtime 을 생산 worker 와 대조한다 — desired/effective 는 태그 시험이 활성화로 잰다.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 83:2 | 골든 자기 모순(worker_count ≠ descriptors 수) → Fatal |
| B2 | if | 87:2 | 생산 worker 수 ≠ 골든 → Fatal |
| B3 | range | 92:2 | 골든 서술자 순회 |
| B4 | if | 95:3 | 열쇠(시장 · 가족 · 레인 · 버전) 드리프트 → Error |
| B5 | if | 101:3 | horizon 드리프트 → Error |
| B6 | if | 108:3 | runtime 드리프트 → Error(로트 A — desired/effective 절 제거 뒤 남은 대조) |

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
| `string` | 108:6 |
| `got.Runtime` | 108:13 |
| `t.Errorf` | 109:4 |
| `got.Runtime` | 109:85 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
