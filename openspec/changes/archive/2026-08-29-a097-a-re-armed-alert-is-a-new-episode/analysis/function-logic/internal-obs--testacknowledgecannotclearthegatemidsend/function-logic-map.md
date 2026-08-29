# Function Logic Map: `TestAcknowledgeCannotClearTheGateMidSend`

- Source: `internal/obs/a096_one_send_per_condition_test.go`
- AST evidence: `ast.json` (6 branches)
- Risk scan: `risk-pattern-report.md`

테스트 함수지만 `check_analysis`가 **수정된 기존 Go 함수**로 잡는다. a097이 본문을 바꿨다.

## 무엇을 바꿨고 무엇을 못 바꿨나

바꾼 것: 판정 수단이 **시간에서 원장 효과로** 옮겨졌다. 이전 판은
`select { case <-done: 실패; case <-time.After(50ms): 통과 }`였다 — 부하가 걸린 기계를
정상으로 읽고, 뮤텍스가 사라진 뒤에도 계속 그렇게 읽는다.

지금은 전송이 `Publish` 안에 멈춰 있는 동안 그 행을 journal에서 직접 읽는다. 끼어든
`Acknowledge`는 `AcknowledgeAlert`(`WHERE state = PENDING`)로 그 행을 도장 찍었을
것이므로, **여전히 PENDING이고 여전히 미도장인 행**은 지속 시간이 아니라 사건이다.

못 바꾼 것: **이것도 증명이 아니다.** 확인하는 goroutine이 경합 지점에 *도달했는지*를
관측하는 수단이 없다. 스케줄러가 그것을 한 번도 돌리지 않으면 배제와 같은 값이 읽힌다.

결정론을 얻으려면 프로덕션에 seam이 필요하다. `Notifier.Journal`은 구체 타입
(`*journal.Journal`)이라 호출을 가로챌 수 없고, 이 파일은 외부 테스트 패키지(`obs_test`)라
미노출 훅에 닿지 못한다. **그래서 a097은 이 잠금을 미검증으로 기록하고 넘긴다**
(tasks §8). 개선을 완결로 세지 않는 것이 spec의 마지막 요구사항이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pub` | `blockingPublisher` — 첫 호출에서 park | 같은 파일 | `entered`가 닫히면 전송 진행 중 |
| `j` | `a096Notifier`가 만든 격리 journal | `t.TempDir()` | 조회 실패는 B1 |
| `released` | `atomic.Bool`. release 직전에 true | 테스트 | 순서 위반은 B5 |
| `ackRaced` | Acknowledge 반환 시점의 `!released` | 테스트 goroutine | — |

**불변식**: `<-pub.entered` 이후 전송은 `Publish` 안에 있고, 잠금이 제 역할을 하면 `n.mu`를
쥐고 있다. 그 구간에서 그 행은 PENDING이다 — `MarkAlertDelivered`는 `Publish` 반환 뒤다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@394 | `PendingAlerts` 오류 | `t.Fatalf` | 중단 | 장애 경로 (`not-applicable`) |
| B2@397 | 대기 행이 1이 아니다 | `t.Fatalf` | 중단 | 배제 실패 시 진입 |
| B3@401 | `AcknowledgedBy != ""` | `t.Errorf` | — | **배제 실패의 직접 증거** |
| B4@409 | `Notify` 오류 | `t.Fatalf` | 중단 | 장애 경로 (`not-applicable`) |
| B5@412 | `ackRaced` — release 전에 반환 | `t.Errorf` | — | 순서 위반의 두 번째 관측 |
| B6@416 | `Acknowledge` 오류 | `t.Fatalf` | 중단 | 장애 경로 (`not-applicable`) |

B2·B3·B5가 판정이고 셋 다 **사건**이다. `time.Sleep(50ms)`는 기회 제공이며 어떤 단언의
근거도 아니다 — 그것이 이전 판과의 차이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Notify` (goroutine) | 전송을 in-flight로 만든다 | 오류는 B4 | AST calls |
| `n.Acknowledge` (goroutine) | 배제 대상 | 오류는 B6 | AST calls |
| `j.PendingAlerts` | **원장 효과 관측** | 오류는 B1 | AST calls |
| `released.Store` / `Load` | 순서 관측 | — | AST calls |

## State mutations and fallbacks

- 프로덕션 상태 변경 없음. journal은 `t.TempDir()`에 격리된다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 테스트 본문 전체.
- High-risk impact: no (테스트)
- 이 테스트가 지키려는 프로덕션 한 줄: `Acknowledge`의 `n.mu.Lock()`.
  **지키지 못한다고 기록한다** — 뮤테이션으로 재현 가능한 탐지를 만들지 못했다.
