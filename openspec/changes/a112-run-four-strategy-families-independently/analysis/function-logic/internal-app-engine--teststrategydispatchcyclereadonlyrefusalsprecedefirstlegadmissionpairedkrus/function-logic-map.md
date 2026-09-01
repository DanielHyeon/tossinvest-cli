# Function Logic Map: `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go`
- Function: `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` in package `engine`
- Signature: `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS(params=1, results=0)`
- File SHA-256: `ed412100d736dfcb474a0b6c126379c383dc0495be9bdceb409808c18d76f844`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 태스크 5.5-fix3(봉투 타입).

## 이 로트가 이 함수에서 바꾼 것

읽기 전용 거절(보호·진입 게이트)이 q_final 승인보다 먼저 일어나 캠페인이 손대지 않은 채 남는지 잰다.
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
| `t.Run` | 154:4 |
| `string` | 154:22 |
| `pairedStrategyDispatchCycleFixture` | 155:33 |
| `entries.authority.Proposal` | 156:15 |
| `proposals.forMarket` | 156:15 |
| `errors.New` | 157:16 |
| `strings.ToLower` | 159:44 |
| `string` | 159:60 |
| `strings.ToLower` | 161:43 |
| `string` | 161:59 |
| `cycle.dispatch` | 163:18 |
| `context.Background` | 163:33 |
| `deliverForTest` | 163:55 |
| `errors.Is` | 163:84 |
| `t.Fatalf` | 164:6 |
| `j.CurrentPositionCampaignCAS` | 166:17 |
| `context.Background` | 166:46 |
| `string` | 167:6 |
| `t.Fatalf` | 169:6 |
| `spy.mu.Lock` | 171:5 |
| `len` | 172:15 |
| `spy.mu.Unlock` | 173:5 |
| `t.Fatalf` | 175:6 |

- AST 호출 23개. 전체 열거는 `ast.json` 에 있다.
- production 심볼과의 결합은 `cycle.dispatch` 와 픽스처가 만드는 권한 값들이며,
  이 로트가 더한 유일한 결합은 `deliverForTest` 다.

## State mutations and fallbacks

- 임시 저널에만 쓴다. 브로커·설정·운영 토글에 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 시험 전용 함수다. production 동작을 만들지 않는다.
- High-risk impact: no — 실계좌 주문 경로에 닿지 않는다(spy Gateway).
