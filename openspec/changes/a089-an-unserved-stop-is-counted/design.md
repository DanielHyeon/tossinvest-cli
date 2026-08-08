# a089 설계 — 안 나간 손절을 세고, 남기고, 보이게 한다

## 척도를 시간에서 횟수로 바꾼다

첫 판은 "9분 무보호"를 고치려 했다. 그 9분은 존재하지 않았다 — 관측 순간의 이탈은
간헐적이었고, 관측 사이의 가격은 원장에 없다. 확실한 것은 **엔진이 손절을 여섯 번
결정했고 다섯 번 나가지 못했다**는 것이다.

그래서 이 change의 일차 척도는 **연속 미제출 횟수**다. 시간은 이차 척도로 남긴다.

## D1 — 불변식 하나로 통일한다

지금 `noteDelay`는 두 곳에서 불리고 `clearDelay`는 한 곳에서 불린다. 세 지점이 서로
다른 것을 뜻하고, 그래서 서로를 무효화한다.

바꿀 것은 호출 개수가 아니라 **의미**다.

> **지연 시계는 "보호 청산이 필요한데 살아 있는 보호 주문이 없다"는 상태가 지속된 시간을
> 잰다.** 시작은 그 상태에 들어갈 때, 해제는 그 상태를 벗어날 때다.

이 정의에서 세 결론이 곧바로 나온다.

1. **`clearDelay`(1150)는 틀렸다.** 그것은 "심볼에 걸리적거리는 주문이 없다"를 뜻하고,
   보호 주문이 살아났다는 뜻이 아니다. 오히려 그 직후에 제출이 실패한다.
2. **해제는 `submit`의 `StateConfirmed`와 포지션 종료뿐이다.**
3. **`StateInDoubt`는 해제하지 않는다.** 주문이 존재하는지 모르는 상태이고, 모르는 것을
   보호로 셀 수 없다. 시작도 하지 않는다 — 해소는 대사기의 몫이고 그 경로에 자체 알림이 있다.

### 지금 이 상태로 끝나는 경로 전부

주문 가능한 **보호** 제안이 살아 있는 주문 없이 주기를 끝내는 지점이다. 전수다.

| # | 위치 | 지금 하는 일 | 시계 |
| --- | --- | --- | --- |
| P1 | `judge:1145` working order를 못 치웠다 | `noteDelay` + `ArmSuppressedWorkingOrder` | ✅ |
| P2 | `judge:1171` `ErrProposalPending` | `return nil` | ❌ 침묵 |
| P3 | `judge:1190` `ArmOutcome != Armed` | `return nil` | ❌ 침묵 |
| P4 | `submit:1243` 확정 floor가 0 | `release` | ❌ 침묵(알림도 없음) |
| P5 | `submit:1263` `IssueReduction` 실패 | `alertProposalRefused` | ❌ |
| P6 | `submit:1276` `sellIntent` 실패 | `alertRefused` | ❌ |
| P7 | `submit:1301` `ReasonSymbolInFlight` | `noteDelay` | ✅ |
| P8 | `submit:1304` 브로커 거부 | `alertProposalRefused` | ❌ ← 8/5 사건 |
| P9 | `submit:1272` `AttachExitIntent` 실패 | 오류 반환 | ❌ (반환은 보인다) |

여덟이 아니라 아홉이고, 계측된 것은 둘이다. 첫 판은 P8 하나만 봤다.

**구현은 아홉 곳에 한 줄씩 넣는 것이 아니다.** 위 불변식을 한 지점에서 판정한다 —
주기의 끝에서 "보호가 필요한가 / 살아 있는 보호 주문이 있는가"를 묻고 시계를 갱신한다.
분기마다 손으로 넣으면 열 번째 경로가 생길 때 또 빠진다.

### 받아들이는 결과

해제를 주문의 존재에 묶으면, 손절이 한 번 거부된 뒤 가격이 회복해 며칠 뒤 익절로 끝나는
포지션도 그 사이 **한 번** 알림을 낸다(`delayAlerted` latch가 반복을 막는다).

a074가 정확히 이 위험을 주석에 적어 뒀다.

> `exitloop.go:1136` — "a genuine clear failure days later would alarm immediately with an
> elapsed time of days."

그 주석의 대상은 **보류된 익절**이고, 익절은 지금도 시계에 들어가지 않는다(P1~P9는 보호
한정). 보호에 대해서는 "며칠 전에 결정된 손절이 아직 안 나갔다"가 **옳은 알림**이다.
더 시끄러운 방향이므로 채택한다. tasks 5.1의 실측 재생이 이 결과를 확인한다.

## D2 — 연속 미제출을 센다

포지션마다 두 값을 유지한다.

- **연속 미제출 횟수** — P1~P9 중 어느 경로로든 보호 제안이 주문이 되지 못하면 +1
- **최초 미제출 시각** — 그 연속의 시작. 시계의 시작과 같은 값이다

**초기화는 접수(`StateConfirmed`)와 포지션 종료뿐이다.** 시계의 해제 조건과 동일하다 —
두 값은 같은 상태의 두 표현이고, 따로 관리하면 어긋난다.

이 값이 `exit.proposal_refused`와 `exit.liquidation_delayed`의 필드로 나간다. 8/5라면
마지막 거부가 `consecutive_unserved: 5`를 달고 나갔을 것이다.

### 어디에 두는가

`ExitObserver`의 기존 맵과 같은 자리, **프로세스 메모리**다(`delayedSince`·`delayAlerted`·
`refused`가 이미 그렇다). 스키마를 올리지 않는다 — 이 저장소는 SchemaVersion 불일치로
엔진이 조용히 죽은 적이 있고, 계측을 위해 그 위험을 살 이유가 없다.

**결과: 재시작하면 카운터와 시계가 0으로 돌아간다.** 이것은 결함이 아니라 명시된 성질이다.
재시작 직후의 포지션은 살아 있는 보호 주문이 없으면 곧바로 다시 세기 시작한다.
원장에 남는 것은 알림과 이벤트이고, 그것은 재시작을 넘어 보존된다.

## D3 — outbox 재발을 한 함수에서 고친다

증상은 넷인데 원인은 하나다. `EnqueueAlert`가 재발을 **행의 상태와 무관하게** 기존 id로
반환하고(`outbox.go:128-131`), 그 뒤의 모든 갱신이 `state = 'PENDING'`을 요구한다.

수리는 **`EnqueueAlert` 한 곳**이다. 재발을 만나면 그 행을 다시 `PENDING`으로 열고
제목·본문·payload를 최신 발생으로 갱신한다. 나머지 함수는 손대지 않는다 — 술어가 다시
만족되므로 `MarkAlertDelivered`·`MarkAlertAttemptFailed`·`AcknowledgeAlert`가
있던 그대로 동작한다.

네 증상이 함께 사라진다.

| 증상 | 왜 사라지는가 |
| --- | --- |
| `attempts`가 1에 멈춘다 | `Mark*`의 술어가 만족돼 누적된다 |
| 재발의 전달 실패가 안 세어진다 | 행이 `PENDING`이므로 `UndeliveredCount`에 든다 |
| `Flush`가 재시도하지 않는다 | `PendingAlerts`가 돌려준다 |
| `Acknowledge`가 빈손으로 게이트를 푼다 | `remaining == 0`이 더는 자명하지 않다 |

**알림 발송량은 변하지 않는다.** `notifyCritical`이 중복이든 아니든 항상 `deliver`를
부르므로(`notifier.go:177-182`) 발송은 지금도 매 발생마다 나간다. 이 변경은 **장부를
발송에 맞추는 것**이지 새 발송을 만드는 것이 아니다. 8/5의 다섯 발송이 그 증거다.

### `ACKNOWLEDGED` 행도 다시 여는가 — 연다

이것이 이 change에서 가장 공격받을 결정이므로 근거를 적는다.

운영자가 1번째 발생을 확인했다는 것은 **5번째 발생을 봤다는 증거가 아니다.**
outbox.go 머리 주석이 대칭인 문장을 이미 적어 뒀다.

> "release is an explicit operator acknowledgement, because 'the network came back' is not
> evidence that anybody read the alert."

'네트워크가 돌아왔다'가 사람이 아니듯, '한 시간 전에 확인했다'는 지금 발생한 것에 대한
확인이 아니다. 그러므로 재발은 다시 열린다.

게이트가 영영 안 풀릴 위험은 없다. 게이트는 **전달이 소진됐을 때만** 잠기고
(`notifier.go:283`), 전달이 정상이면 행은 같은 호출 안에서 `PENDING → DELIVERED`로
돌아간다. 전달이 깨져 있고 조건이 계속 재발하는 동안 게이트가 잠겨 있는 것은 정확히
의도된 동작이다.

## D4 — 브로커의 사유를 옮겨 적는다

`official.APIError.Body`의 JSON에서 `error.code`와 `error.data.field`를 꺼내 별도 필드로
남긴다. 가산 메서드 하나이고 기존 동작·타입·`ShouldFallback`은 그대로다.

**분류하지 않는다.** 표를 만들지 않고, 열거형을 만들지 않고, 사유에 따라 대응을 바꾸지
않는다. 옳은 대응을 정하려면 각 사유의 실제 결과를 알아야 하는데 그 측정이 없다.
있는 것은 브로커가 준 문자열뿐이고, 이 change는 그것을 **질의 가능한 자리로 옮길** 뿐이다.

없으면 빈 값이다. 파싱이 실패해도 기존 `detail` 원문은 그대로 남으므로 잃는 정보가 없다.

§0.8: 옮기는 두 값은 `code`와 `field`이며 계좌·세션·개인정보를 담지 않는다. 같은 본문이
이미 `detail`로 로그에 남고 있다.

## 건드리지 않는 것

- **판정기 `exitpolicy`** — 순수 함수이고 그 판정은 옳다. 8/5의 여섯 제안은 전부 정확했다
- **재발의 주기** — `Changed`가 `|| s.Orderable`로 끝나므로 이미 매 주기 보장된다
- **주문의 가격·유형** → a087
- **`sellIntent`·`execgw`·`internal/trading`** — 이 change는 주문을 만들지 않는다
- **스키마** — 올리지 않는다
- **`notifier.go`의 전달·게이트·확인 로직** — D3은 `outbox.go` 한 함수다

## 검증

- 시계: 아홉 경로 각각에서 시작/지속되고, `StateConfirmed`와 포지션 종료에서만 해제된다
- 시계: `clearTheSymbol` 성공만으로는 **해제되지 않는다**(C1 회귀 테스트)
- 시계: `StateInDoubt`는 시작도 해제도 하지 않는다
- 시계: 한계 초과 시 critical 1회, `delayAlerted` latch로 반복 없음
- 카운터: 아홉 경로에서 +1, `StateConfirmed`에서 0, 종료에서 삭제
- 카운터: 알림 필드에 실린다
- outbox: `DELIVERED` 행 재발 → `PENDING` 복귀 + 내용 갱신 + `attempts` 누적
- outbox: `ACKNOWLEDGED` 행 재발 → `PENDING` 복귀
- outbox: 재발의 전달 실패가 `UndeliveredCount`에 든다
- outbox: 첫 발생 동작은 무변화(회귀)
- 사유: `code`·`field`가 필드로 나가고, 파싱 실패 시 빈 값이며 `detail`은 보존된다
- **실측 재생**: 8/5의 여섯 판정·다섯 거부를 fixture로 재생 —
  알림이 00:55:02에 경과 62초로 한 번, 마지막 거부가 `consecutive_unserved: 5`,
  outbox 행의 `attempts`가 5
- `go test ./... -count=1 -race` 회귀 0, upstream 650 green 유지
