# Function Logic Map: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go`
- Source SHA-256: `b0b9734d75c5e4fafa2b2033bd9c3d9660af813aa7d485b1840d2a1f8ebcd958`
- Signature: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner(params=1, results=0)`
- Source range: `97:1`–`167:2`
- AST evidence: `ast.json` — 현재 소스(a112 게이트 위생 재측정 2026-10-01).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 입력은 `*testing.T` 하나. spy Gateway 와 임시 저널만 쓰고 실계좌에 닿지 않는다.
- 주제(그대로): 두 dispatch 가 동시에 출발하고 하나의 중앙 owner 아래에서 돌며 번갈아 기다리지 않는다. 공유 bucket 두 번째 진입은 BUCKET_USAGE_STALE 로 거절된다(5.6.1 계약).

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 115:2 | KR · US 두 dispatch 를 동시에 출발시키는 순회 |
| B2 | range | 133:2 | 결과 둘을 받는 순회 |
| B3 | switch | 134:3 | 결과 분류 |
| B4 | case | 135:3 | 성립(Confirmed) — admitted 에 담음 |
| B5 | case | 137:3 | 공유 bucket 두 번째 진입의 `ATOMIC_ADMISSION_FAILED`(BUCKET_USAGE_STALE) — refused 에 담음 |
| B6 | case | 143:3 | 그 밖의 결과 — 실패 단언 |
| B7 | if | 147:2 | 성립 하나 · 거절 하나 · 서로 다른 시장이 아니면 실패 |
| B8 | if | 153:2 | Gateway 호출이 성립한 시장 하나가 아니면 실패 |
| B9 | if | 157:2 | lease 읽기 실패 |
| B10 | if | 163:2 | lease 의 owner epoch · fencing token 이 중앙 owner(첫 epoch)와 다르면 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `pairedStrategyDispatchCycleFixture` | 106:30 |
| `make` | 112:13 |
| `make` | 113:11 |
| `deliverForTest` | 120:16 |
| `entries.authority.Proposal` | 120:34 |
| `proposals.forMarket` | 120:34 |
| `runners.Add` | 121:3 |
| `(unnamed)` | 122:6 |
| `runners.Done` | 123:10 |
| `cycle.dispatch` | 125:16 |
| `context.Background` | 125:31 |
| `close` | 129:2 |
| `runners.Wait` | 130:2 |
| `close` | 131:2 |
| `append` | 136:15 |
| `strings.Contains` | 137:29 |
| `result.err.Error` | 137:46 |
| `strings.Contains` | 138:4 |
| `result.err.Error` | 138:21 |
| `append` | 142:14 |
| `t.Fatalf` | 144:4 |
| `len` | 147:5 |
| `len` | 147:27 |
| `t.Fatalf` | 148:3 |
| `spy.mu.Lock` | 150:2 |
| `append` | 151:11 |
| `(unnamed)` | 151:18 |
| `spy.mu.Unlock` | 152:2 |
| `len` | 153:5 |
| `strings.EqualFold` | 153:25 |
| `string` | 153:67 |
| `t.Fatalf` | 154:3 |
| `j.LookupStrategyDispatchLease` | 156:16 |
| `context.Background` | 156:46 |
| `t.Fatal` | 158:3 |
| `t.Fatalf` | 164:3 |

## State mutations and fallbacks

- 임시 저널에 dispatch lease · 예약을 쓴다(시험 소유). 생산 상태 없음.

## Safety conclusion

- High-risk 아님(시험 코드). 거절된 쪽이 Gateway 에 닿지 않음을 단언한다(B8).
