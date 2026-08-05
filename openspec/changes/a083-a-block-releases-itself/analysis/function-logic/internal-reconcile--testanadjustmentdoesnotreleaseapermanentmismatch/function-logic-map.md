# Function Logic Map: `TestAnAdjustmentDoesNotReleaseAPermanentMismatch`

- Source: `internal/reconcile/mismatch_test.go` (lines 243–269)
- AST evidence: `ast.json` (`source_sha256: 797c553eb4d9cc2c161b1a7e6bb6cc78f6f77d4a3a38e9c9b768061466bf61a1`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

영구 불일치는 credit으로 풀리지 않는다. a083에서 as-of를 붙이고 관측 전에 시계를 진행시켜, 거부하는 것이 stamp가 아니라 영구성임을 분명히 했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (248) `for` — for i := 0; i < 3; i++ | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (252) `if` — if !tracker.Permanent() | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (262) `if` — if !tracker.Permanent() | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (266) `if` — if rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 244, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 245, 'column': 10}, 'text': 'execgw.NewEntryGate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 246, 'column': 13}, 'text': 'newTracker'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 249, 'column': 23}, 'text': 'mismatchDiffAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 249, 'column': 3}, 'text': 'observe'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 250, 'column': 3}, 'text': 'clk.Advance'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 252, 'column': 6}, 'text': 'tracker.Permanent'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 253, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 258, 'column': 28}, 'text': 'asOfAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 258, 'column': 2}, 'text': 'tracker.AdjustmentApplied'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 259, 'column': 2}, 'text': 'clk.Advance'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 260, 'column': 22}, 'text': 'cleanDiffAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
