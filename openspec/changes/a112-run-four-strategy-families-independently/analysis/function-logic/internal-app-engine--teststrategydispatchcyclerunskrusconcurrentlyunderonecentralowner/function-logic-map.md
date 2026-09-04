# Function Logic Map: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go`
- Function: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` in package `engine`
- Signature: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner(params=1, results=0)`
- File SHA-256: `10f38fe4e88c3e076e8c88b1cc4764c847fb6052835e3be2595c85f34e9b1464`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 9.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 태스크 5.5-fix3(봉투 타입).

## 이 로트가 이 함수에서 바꾼 것

두 시장의 dispatch 가 동시에 돌아도 소유자 epoch 와 fencing token 이 갈라지지 않는지 잰다.
바뀐 것은 **dispatch 에 넘기는 값을 만드는 방법 하나**다.

앞 판본은 `cycle.dispatch(ctx, result)` 처럼 경계를 지나지 않은
`strategyflow.Result` 를 곧바로 넘겼다. `strategyDispatchCycle.dispatch` 가 이제
`strategyhandoff.Delivered` 를 받으므로 그 철자는 **컴파일되지 않는다**. 그래서
시험도 생산 코드와 같은 문을 지난다 — `deliverForTest` 가
`strategyhandoff.Admit(true, …).Deliver(…)` 로 봉투를 만든다
(`strategy_dispatch_envelope_test.go`).

시험만 쓰는 뒷문을 만들지 않은 것이 요점이다. 뒷문을 만들면 그 뒷문이 곧 생산
코드가 쓸 수 있는 문이 된다.

봉투는 고루틴 **밖에서** 미리 만든다. 안에서 만들면 `t.Fatalf` 가 시험 고루틴 밖에서
불릴 수 있고, 두 dispatch 가 같은 순간에 출발한다는 이 시험의 요점도 흐려진다 —
경주 전에 하는 일이 늘어나기 때문이다.

**재는 것은 달라지지 않았다.** 단언·기대값·호출 순서는 그대로이고, 이 시험이
초록인 것은 앞 판본과 같은 이유다.

## Inputs and invariants

- 입력은 위 signature 그대로다(`*testing.T` 하나).
- 불변식: 이 시험은 spy Gateway 와 임시 저널만 쓰고 실계좌에 닿지 않는다.

## Branches and early returns

- AST 분기 9개, 반환 0개. 전체 열거는 `ast.json` 에 있다.
- 분기는 전부 단언 실패 경로(`t.Fatalf`)와 시장 순회다. production 분기를 만들지 않는다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `pairedStrategyDispatchCycleFixture` | 93:30 |
| `make` | 99:13 |
| `make` | 100:11 |
| `deliverForTest` | 107:16 |
| `entries.authority.Proposal` | 107:34 |
| `proposals.forMarket` | 107:34 |
| `runners.Add` | 108:3 |
| `(unnamed)` | 109:6 |
| `runners.Done` | 110:10 |
| `cycle.dispatch` | 112:16 |
| `context.Background` | 112:31 |
| `close` | 116:2 |
| `runners.Wait` | 117:2 |
| `close` | 118:2 |
| `t.Fatalf` | 122:4 |
| `t.Fatalf` | 127:3 |
| `spy.mu.Lock` | 129:2 |
| `append` | 130:11 |
| `(unnamed)` | 130:18 |
| `spy.mu.Unlock` | 131:2 |
| `len` | 132:5 |
| `t.Fatalf` | 133:3 |
| `j.LookupStrategyDispatchLease` | 138:17 |
| `context.Background` | 138:47 |
| `t.Fatal` | 140:4 |
| `t.Fatalf` | 146:4 |

- AST 호출 26개. 전체 열거는 `ast.json` 에 있다.
- production 심볼과의 결합은 `cycle.dispatch` 와 픽스처가 만드는 권한 값들이며,
  이 로트가 더한 유일한 결합은 `deliverForTest` 다.

## State mutations and fallbacks

- 임시 저널에만 쓴다. 브로커·설정·운영 토글에 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 시험 전용 함수다. production 동작을 만들지 않는다.
- High-risk impact: no — 실계좌 주문 경로에 닿지 않는다(spy Gateway).
