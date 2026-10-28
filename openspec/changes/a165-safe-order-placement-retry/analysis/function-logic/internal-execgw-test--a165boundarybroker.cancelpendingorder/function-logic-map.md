# Function Logic Map: `a165BoundaryBroker.CancelPendingOrder`

- Source: `internal/execgw/a165_cancel_refusal_boundary_test.go` (40-51)
- Revision: D5 4판 GREEN, `source_sha256` `e728c257f29ffaedbf547080c55c64c836b900c8859895bd4563a1d2d276883e`
- AST evidence: `ast.json` — AST branches 2, returns 2, calls 2.
- Risk scan: `risk-pattern-report.md`, configured pattern 발견 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `calls` | 1 이상 | fake 내부 카운터 | 첫 호출만 AP 409를 반환함 |
| `firstDelay` | 0 이상 | boundary test fixture | 실제 sleep은 Fake Clock을 전진시키지 않음 |
| retry `ctx` | window context | wrapper | deadline 여부만 관찰하고 ACK 반환함 |

## Branches and early returns

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if at 42:2` | 첫 cancel | calls 증가 | AP 409 경로 | `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement` |
| B2 | `if at 43:3` | `firstDelay > 0` | 실제 sleep | AP 409 반환 | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Sleep` | 첫 broker 왕복 지연 모사 | Fake Clock과 독립 | AST 44 |
| `ctx.Deadline` | retry context 전달 관찰 | deadline 값만 기록 | AST 49 |

## State mutations and fallbacks

- `calls`, `secondDeadline`만 fixture에 기록함. 외부 상태·journal mutation 없음.
- 두 번째 호출은 ACK를 반환해 wrapper의 retry 결과 경계를 검증함.

## Safety conclusion

- Safe edit boundary: D5 test fixture 관찰 전용임.
- High-risk impact: no — production path가 아니나 실시간 지연과 Fake Clock 분리 계약을 고정함.
