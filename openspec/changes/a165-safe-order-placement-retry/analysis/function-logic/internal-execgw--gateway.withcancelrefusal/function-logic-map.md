# Function Logic Map: `Gateway.withCancelRefusal`

- Source: `internal/execgw/cancel_refusal.go` (44-102)
- Qualified function: `Gateway.withCancelRefusal`
- Revision: D5 4판 GREEN, `source_sha256` `fd04d50f9a03ee097a0e1a0b336d0b1faaf3794d8e061c24820cca3c780b2ebf`
- AST evidence: `ast.json` — AST 분기 12, 반환 12, 호출 18(`go run ./tools/logic-map`, 2026-10-29).
- Risk scan: `risk-pattern-report.md`. 직접 `time.Now`/`time.Since`/`context.WithDeadline` 호출 0. `context.WithTimeout(dctx, policy.Window)` 하나는 매달린 재전송 상한일 뿐이고 판정은 `g.clk`만 사용함.

## 역할

취소의 409 `already-processing` 결과만 원 dispatch 바깥에서 경로별 정책으로 처리함. 엔진은 같은 주기 대기·재전송 없이 확정 미수락으로 종결하고, flatten은 누적 5초·최대 2회·8초 창 안에서만 재전송하며 소진·중단은 첫 AMBIGUOUS로 되돌림.

## Inputs and invariants

| 입력/상태 | 정본 | 실패 방향 |
|---|---|---|
| `g.cancelRefusal`, `plan.kind` | Gateway 배선 | nil 또는 취소 외에는 받은 dispatch를 바이트 동일하게 반환함 |
| `g.clk` | 주입 Clock | 창 판정은 `Now`/`Since` 한 시계만 사용함 |
| `dctx` | 호출자 context | 재전송 상한 context의 부모이며, 취소되면 첫 AMBIGUOUS로 fail-closed 정착함 |
| `policy` | engine/flatten 생성자 | 엔진의 대기·재전송은 0, flatten만 제한된 재전송 허용함 |

## Branches and early returns

| 분기 | 좌표 | 조건 | 결과 | 시험 |
|---|---|---|---|---|
| B1 | 45 | 정책 nil 또는 취소 외 | 원 dispatch 반환 | `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` |
| B2 | 53 | 첫 결과가 AP 409 AMBIGUOUS 아님 | 첫 결과 보존 | `TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch` |
| B3 | 56 | definitive 엔진 정책 | REJECTED/FAILED_CONFIRMED, 재전송 0 | `TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay` |
| B4 | 61 | flatten 정책 값 무효 | 오늘의 첫 결과 | 정책 정확값 시험 |
| B5 | 70 | 재전송 loop | 최대 Retries회 | exhaustion 시험 |
| B6 | 72 | 누적 대기 상한 | 첫 AMBIGUOUS + stop=max_wait | `TestA165FlattenRetriesWithinCumulativeBudget` |
| B7 | 75 | sleep 전 Clock 창 경계 | 첫 AMBIGUOUS + stop=window | `TestA165WindowStartsBeforeFirstDispatch` |
| B8 | 78 | sleep context 종료 | 첫 AMBIGUOUS + stop=cancelled | `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled` |
| B9 | 85 | sleep 뒤 Clock 창 만료 | 첫 AMBIGUOUS | 경계 시험 |
| B10 | 85 | 재호출 결과가 AP 아님 | NOT_SENT 또는 ACK 분기 | `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome` |
| B11 | 92 | retry NOT_SENT | 첫 결과 + stop=not_sent | `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome` |
| B12 | 93 | retry ACK | 재시도 결과 + stop=ack | `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement` |

## D5 4판 구현 결과

- 삭제: `wallStart := time.Now()` 및 `context.WithDeadline(dctx, wallStart.Add(policy.Window))`.
- 추가: `context.WithTimeout(dctx, policy.Window)`. 이것은 판정 시계가 아니라 매달린 재전송을 끊는 실시간 상한임. 생산에서는 systemClock과 같은 진행이라 Clock 판정과 일치하고, Fake Clock 시험에서는 Clock 판정이 먼저 끝나도록 시험함.
- 실시간 상한이 먼저 끝나면 `Sleep` 오류가 `stop=cancelled outcome=today`로 첫 409 AMBIGUOUS를 보존함. 더 덜 안전한 갈래는 없음.
- AST 핀: 이 파일에서 `time.Now`·`time.Since` 0, `context.WithTimeout` 1, `context.WithDeadline` 0을 고정함.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.clk.Now`/`g.clk.Since` | 창 판정 | 주입 Clock만 사용함 | AST 50·75·81 |
| `context.WithTimeout` | 매달린 retry 상한 | 부모는 `dctx`, 판정에 쓰지 않음 | AST 66 |
| `g.clk.Sleep` | retry-after 대기 | 취소되면 첫 결과 보존 | AST 79 |
| `dispatch` | 원 cancel call | 첫 호출 `dctx`, retry `winCtx` | AST 51·89 |

## State mutations and fallbacks

- 직접 journal mutation 없음. 결과 class/detail만 반환함.
- retry 상한·호출자 취소·NOT_SENT·소진은 모두 첫 409 AMBIGUOUS를 보존함.

## Safety conclusion

Clock 창 판정과 전송 취소 상한을 구분해도 재전송 허용 조건은 `g.clk` 하나에서만 결정됨. 부모 `dctx`를 유지하므로 호출자 취소를 무시하지 않고, 실시간 상한은 첫 AMBIGUOUS를 IN_DOUBT 쪽으로 보존함. High-risk impact: yes.
