# Function Logic Map: `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement`

- Source: `internal/execgw/a165_cancel_refusal_boundary_test.go` (85-107)
- Revision: D5 4판 GREEN, `source_sha256` `e728c257f29ffaedbf547080c55c64c836b900c8859895bd4563a1d2d276883e`
- AST evidence: `ast.json` — AST branches 4, returns 0, calls 11.
- Risk scan: `risk-pattern-report.md`, configured pattern 발견 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `firstDelay` | 실제 25ms | fixture | Clock 창 판정으로 쓰면 안 됨 |
| Fake Clock | 0에서 retry-after만 전진 | `a165AdvancingClock` | 1ms가 아니면 판정 경계 오류 |
| `Window` | 10ms | 정책 | retry 확인은 주입 Clock 결과를 따름 |

## Branches and early returns

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if at 94:2` | cancel 결과 | 없음 | CONFIRMED 요구 | 함수 자신 |
| B2 | `if at 97:2` | broker calls | 없음 | 2회 요구 | 함수 자신 |
| B3 | `if at 101:2` | journal record | 없음 | CONFIRMED 요구 | 함수 자신 |
| B4 | `if at 104:2` | Fake Clock elapsed | 없음 | 1ms 요구 | 함수 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a165BoundaryCancel` | production wrapper 호출 | AP 뒤 retry를 관찰 | AST 93 |
| `j.LookupAttempt` | 정착 상태 확인 | CONFIRMED가 아니면 실패 | AST 100 |
| `clk.Since` | 판정 시계 확인 | 실제 지연과 분리 | AST 104 |

## State mutations and fallbacks

- fixture journal에 cancel attempt 하나를 기록함.
- 실패는 test fatal이며 production fallback을 변경하지 않음.

## Safety conclusion

- Safe edit boundary: D5 시험 기대값임.
- High-risk impact: yes — wall clock이 창 판정으로 되돌아가는 회귀를 막음. 실시간 상한 갈래는 `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled`가 첫 AMBIGUOUS/`stop=cancelled outcome=today`로 고정함.
