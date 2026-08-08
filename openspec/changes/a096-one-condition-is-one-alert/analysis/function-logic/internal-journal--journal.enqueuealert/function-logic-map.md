# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go` (115–122)
- AST evidence: `ast.json` (sha256 `c6612c641a3a…`, 0분기, 반환 1곳)
- Risk scan: `risk-pattern-report.md`

a096 이전 이 함수는 9분기의 본체였다. 지금은 `ClaimAlertForDelivery`에 위임하고
두 번째 반환값을 버린다. 서명·계약·중복 제거 동작은 **바뀌지 않았다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a` | `EventKey`·`Type` 비공백 | 호출자 | `ClaimAlertForDelivery`가 판정하고 오류를 그대로 올린다 |
| 반환 id | 중복 제거된 행의 id | `alert_outbox.event_key` UNIQUE (`outbox.go:50`) | 같은 key는 같은 id |

불변식: **호출자에게 보이는 동작이 base `ec29dc72`와 동일하다.** 서명도 반환값의 의미도
같으므로 기존 호출자(`execgw.Gateway.parkAlert`, `internal/journal/outbox_test.go`의 7개
테스트)는 **한 글자도 바뀌지 않았다**. 이것이 이 위임을 택한 이유다 — 원장 패키지의
기존 테스트를 기계적으로 고치면 그 테스트들이 전부 "수정된 기존 함수"가 되고,
그때 생산되는 증거는 각 테스트의 안전성이 아니라 arity 변경의 기록일 뿐이다.

## Branches and early returns

분기가 없다. 위임 한 줄과 반환 한 줄이다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (분기 없음 — 유일 경로) | 없음 (위임, `remindAfter = 0`) | `ClaimAlertForDelivery`의 id와 오류 | `TestEnqueueAlertKeepsItsContract` + 기존 7개 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.ClaimAlertForDelivery` | 기록과 중복 제거의 실제 구현 | 오류를 그대로 올린다 | AST :120 |

호출자(CodeGraph): `execgw.Gateway.parkAlert`
([replay.go:551](../../../../../internal/execgw/replay.go#L551), 반환값을 버린다)와
`internal/journal/outbox_test.go`. `obs.Notifier.notifyCritical`은 a096 이후
**이 함수를 부르지 않는다** — `ClaimAlertForDelivery`를 직접 부른다.

`parkAlert`가 이 함수를 계속 쓰는 것은 타협이 아니다: replay 경로는 알림을 **기록만** 하고
전송 루프를 갖지 않으므로, 전달이 아직 필요한지는 그 호출자가 답할 수 있는 질문이 아니다.
그래서 `remindAfter = 0`을 넘긴다 — 남의 리마인더 정책을 대신 실행하지 않는다.

`replay.go:101`의 주석은 그 행을 `Notifier.Flush`가 집어 간다고 말하지만 **`Flush`에는
non-test 호출자가 없다**(독립 리뷰 1라운드 concern 3, 확인함). a096이 만든 문제도 고치는
문제도 아니며 tasks 7.4로 남긴다.

## State mutations and fallbacks

- 이 함수 자체는 아무것도 쓰지 않는다. 모든 쓰기는 위임 대상 안의 한 트랜잭션이다.

## Safety conclusion

- Safe edit boundary: 본체를 위임으로 바꾸는 것. 서명·반환 의미·중복 제거가 모두 같으므로
  호출자 관점의 동작 변화가 없다.
- High-risk impact: **yes** — `internal/journal`은 원장이다. 다만 이 편집은 쓰기를 추가하지도
  제거하지도 않고, SQL도 트랜잭션 경계도 그대로 옮겨졌을 뿐이다.
