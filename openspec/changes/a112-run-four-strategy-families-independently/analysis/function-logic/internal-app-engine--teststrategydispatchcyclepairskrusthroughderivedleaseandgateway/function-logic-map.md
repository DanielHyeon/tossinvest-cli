# Function Logic Map: `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go`
- Function: `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway` in package `engine`
- Signature: `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway(params=1, results=0)`
- File SHA-256: `10f38fe4e88c3e076e8c88b1cc4764c847fb6052835e3be2595c85f34e9b1464`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 태스크 5.5-fix3(봉투 타입).

## 이 로트가 이 함수에서 바꾼 것

KR·US 두 시장이 각각 파생 lease 와 Gateway 를 지나 체결까지 가는지 잰다.
바뀐 것은 **dispatch 에 넘기는 값을 만드는 방법 하나**다.

앞 판본은 `cycle.dispatch(ctx, result)` 처럼 경계를 지나지 않은
`strategyflow.Result` 를 곧바로 넘겼다. `strategyDispatchCycle.dispatch` 가 이제
`strategyhandoff.Delivered` 를 받으므로 그 철자는 **컴파일되지 않는다**. 그래서
시험도 생산 코드와 같은 문을 지난다 — `deliverForTest` 가
`strategyhandoff.Admit(true, …).Deliver(…)` 로 봉투를 만든다
(`strategy_dispatch_envelope_test.go`).

시험만 쓰는 뒷문을 만들지 않은 것이 요점이다. 뒷문을 만들면 그 뒷문이 곧 생산
코드가 쓸 수 있는 문이 된다.

**재는 것은 달라지지 않았다.** 단언·기대값·호출 순서는 그대로이고, 이 시험이
초록인 것은 앞 판본과 같은 이유다.

## Inputs and invariants

- 입력은 위 signature 그대로다(`*testing.T` 하나).
- 불변식: 이 시험은 spy Gateway 와 임시 저널만 쓰고 실계좌에 닿지 않는다.

## Branches and early returns

- AST 분기 7개, 반환 0개. 전체 열거는 `ast.json` 에 있다.
- 분기는 전부 단언 실패 경로(`t.Fatalf`)와 시장 순회다. production 분기를 만들지 않는다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `pairedStrategyDispatchCycleFixture` | 64:31 |
| `entries.authority.Proposal` | 65:13 |
| `proposals.forMarket` | 65:13 |
| `cycle.dispatch` | 66:15 |
| `context.Background` | 66:30 |
| `deliverForTest` | 66:52 |
| `t.Fatalf` | 68:4 |
| `len` | 70:6 |
| `t.Fatalf` | 71:4 |
| `j.LookupStrategyDispatchLease` | 74:17 |
| `context.Background` | 74:47 |
| `t.Fatalf` | 77:4 |
| `t.Fatalf` | 82:5 |
| `t.Fatalf` | 88:3 |

- AST 호출 14개. 전체 열거는 `ast.json` 에 있다.
- production 심볼과의 결합은 `cycle.dispatch` 와 픽스처가 만드는 권한 값들이며,
  이 로트가 더한 유일한 결합은 `deliverForTest` 다.

## State mutations and fallbacks

- 임시 저널에만 쓴다. 브로커·설정·운영 토글에 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 시험 전용 함수다. production 동작을 만들지 않는다.
- High-risk impact: no — 실계좌 주문 경로에 닿지 않는다(spy Gateway).
