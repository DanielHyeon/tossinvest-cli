# Function Logic Map: `Gateway.withCancelRefusal`

- Source: `internal/execgw/cancel_refusal.go` (44-102)
- Qualified function: `Gateway.withCancelRefusal`
- Revision: pre-D5 4판 정정, `source_sha256` `576b7752ce9dac5121d61a83f2bdbc5a65c1c71b22b185f4c338e47655586485`
- AST evidence: `ast.json` — AST 분기 12, 반환 12, 호출 18(`go run ./tools/logic-map`, 2026-10-29).
- Risk scan: 직접 `time.Now`/`context.WithDeadline`가 판정 시계와 분리되어 있음. Manager D5 결정에 따라 이 부분만 `context.WithTimeout(dctx, policy.Window)`로 바꾸고 판정은 `g.clk`만 사용함.

## 역할

취소의 409 `already-processing` 결과만 원 dispatch 바깥에서 경로별 정책으로 처리함. 엔진은 같은 주기 대기·재전송 없이 확정 미수락으로 종결하고, flatten은 누적 5초·최대 2회·8초 창 안에서만 재전송하며 소진·중단은 첫 AMBIGUOUS로 되돌림.

## 입력과 불변식

| 입력/상태 | 정본 | 실패 방향 |
|---|---|---|
| `g.cancelRefusal`, `plan.kind` | Gateway 배선 | nil 또는 취소 외에는 받은 dispatch를 바이트 동일하게 반환함 |
| `g.clk` | 주입 Clock | 창 판정은 `Now`/`Since` 한 시계만 사용함 |
| `dctx` | 호출자 context | 재전송 상한 context의 부모이며, 취소되면 첫 AMBIGUOUS로 fail-closed 정착함 |
| `policy` | engine/flatten 생성자 | 엔진의 대기·재전송은 0, flatten만 제한된 재전송 허용함 |

## 분기와 반환

| 분기 | 좌표 | 조건 | 결과 | 시험 |
|---|---|---|---|---|
| B1 | 45 | 정책 nil 또는 취소 외 | 원 dispatch 반환 | `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` |
| B2 | 54 | 첫 결과가 AP 409 AMBIGUOUS 아님 | 첫 결과 보존 | `TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch` |
| B3 | 57 | definitive 엔진 정책 | REJECTED/FAILED_CONFIRMED, 재전송 0 | `TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay` |
| B4 | 62 | flatten 정책 값 무효 | 오늘의 첫 결과 | 정책 정확값 시험 |
| B5 | 70 | 재전송 loop | 최대 Retries회 | exhaustion 시험 |
| B6-B8 | 72-82 | 대기 상한·Clock 창·sleep context | 첫 AMBIGUOUS + stop 꼬리표 | `TestA165WindowStartsBeforeFirstDispatch`, caller cancellation 시험 |
| B9 | 85 | sleep 뒤 Clock 창 만료 | 첫 AMBIGUOUS | 경계 시험 |
| B10-B12 | 92-96 | 재호출 결과 | NOT_SENT는 첫 결과, 성공은 재시도 결과, AP는 다음 loop | NOT_SENT·success·exhaustion 시험 |

## D5 4판 편집 계획

- 삭제: `wallStart := time.Now()` 및 `context.WithDeadline(dctx, wallStart.Add(policy.Window))`.
- 추가: `context.WithTimeout(dctx, policy.Window)`. 이것은 판정 시계가 아니라 매달린 재전송을 끊는 실시간 상한임. 생산에서는 systemClock과 같은 진행이라 Clock 판정과 일치하고, Fake Clock 시험에서는 Clock 판정이 먼저 끝나도록 시험함.
- 실시간 상한이 먼저 끝나면 `Sleep` 오류가 `stop=cancelled outcome=today`로 첫 409 AMBIGUOUS를 보존함. 더 덜 안전한 갈래는 없음.
- AST 핀: 이 파일에서 `time.Now`·`time.Since` 0, `context.WithTimeout` 1, `context.WithDeadline` 0을 고정함.

## 안전 결론

Clock 창 판정과 전송 취소 상한을 구분해도 재전송 허용 조건은 `g.clk` 하나에서만 결정됨. 부모 `dctx`를 유지하므로 호출자 취소를 무시하지 않고, 실시간 상한은 첫 AMBIGUOUS를 IN_DOUBT 쪽으로 보존함.
