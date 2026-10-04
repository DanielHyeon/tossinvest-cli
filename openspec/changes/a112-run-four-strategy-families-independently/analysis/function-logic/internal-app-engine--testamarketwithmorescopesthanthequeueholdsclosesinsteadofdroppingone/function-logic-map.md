# Function Logic Map: `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne (시험)`

- Source: `internal/app/engine/a112_coordinator_test.go`
- Source SHA-256: `60d5adbcf550c3ff422046d010bfe037481c7fa39e5b1561afb5c1d19a1f2115`
- Signature: `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne(params=1, results=0)`
- Source range: `44:1`–`80:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 repin-1e25b3a3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 큐 넘침은 시장을 닫고(항목 0) 버린 수를 세며 중재 코드를 빌리지 않는다 — 관문 아래에서도 같고, 그 닫힘이 판정 활성화를 싣는다.

## Branches and early returns

- Exact AST return nodes: `70:4`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 47:2 | 종목 Capacity+1 개 생성 |
| B2 | if | 51:2 | 미선언 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal |
| B3 | if | 54:2 | 닫힌 시장이 열림 · 항목 있음 → Fatal |
| B4 | if | 57:2 | 버린 수 0 → Fatal(조용한 유실 금지) |
| B5 | if | 61:2 | 넘침이 중재 코드를 빌림 → Fatal |
| B6 | if | 73:2 | **(새)** 관문 아래 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal |
| B7 | if | 76:2 | **(새)** QUEUE_OVERFLOW 닫힘이 판정 활성화를 안 실음 → Fatal |

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
| `familyGateFixture` | 66:25 |
| `collectOverflowing` | 67:11 |
| `loader.withStrategyLanes` | 68:3 |
| `t.Fatalf` | 74:3 |
| `Verified` | 76:6 |
| `gated.kr.familyActivation` | 76:6 |
| `Generation` | 76:48 |
| `gated.kr.familyActivation` | 76:48 |
| `activation.Generation` | 76:92 |
| `t.Fatalf` | 77:3 |
| `Verified` | 78:4 |
| `gated.kr.familyActivation` | 78:4 |
| `Generation` | 78:44 |
| `gated.kr.familyActivation` | 78:44 |
| `activation.Generation` | 78:86 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
