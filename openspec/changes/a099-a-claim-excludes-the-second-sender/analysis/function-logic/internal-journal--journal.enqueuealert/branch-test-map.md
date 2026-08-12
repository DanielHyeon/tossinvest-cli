# Branch Test Map: `Journal.EnqueueAlert`

`ast.json`의 열거가 정본이다: 분기 4 · 이탈 5.
**GREEN 칸은 실측해서 채운다.**

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:133` `alertKey`가 거절한다 — 키 없음 / 타입 없음 | `TestEnqueueAlertRequiresAKeyAndAType` `outbox_test.go:62` | no (기존 통과) | **yes** |
| B2 | `:137` `BeginTx`가 실패한다 | 없음 — 드라이버 오류 주입 없음 | no | **no (기존부터 없다)** |
| B3 | `:147` `recordAlertTx`가 실패한다 | 없음 | no | **no (기존부터 없다)** |
| B4 | `:150` `Commit`이 실패한다 | 없음 | no | **no (기존부터 없다)** |

이탈 `:153`(정상)은 분기가 아니므로 위 표의 행이 아니다. 아래에 따로 적는다.

## 정상 이탈 `:153` — a099가 실제로 관측한 것

| Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|
| 같은 `event_key`를 두 번 넣으면 행 하나 | `TestEnqueueAlertIsIdempotentOnTheEventKey` `outbox_test.go:30` | no (기존 통과) | **yes** |
| **기록 직후 곧바로 청구 가능하다 — D13** | `TestARecordedAlertCanBeClaimedImmediately` `a099_lease_lifecycle_test.go:305` | **yes — §4.7 되돌림 관측**: `EnqueueAlert`를 `ClaimAlertForDelivery` 위임으로 되돌리면 `ClaimHeldElsewhere`가 나온다 | **yes** |
| 경합 시험의 배치가 이 경로를 쓴다 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` `a099_claim_excludes_the_second_sender_test.go:41` | yes (구현 전 두 발송자 모두 발송) | **yes** |

## RED — 시점 관측의 이탈을 여기에도 적는다

§3.2에 적은 대로, §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. 계획은 진행하며 그 자리에서 보는 것이었다.
**되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과 같지 않다.**
그 판정은 리뷰의 몫이다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B2 · B3 · B4** — 드라이버·커밋 실패 경로. a099가 조건도 반환도 안 바꾼다.
  **`not-applicable`: 이 change는 세 분기를 근거로 아무것도 주장하지 않는다.**
  덮여 있지 않다는 사실은 a099가 만든 것이 아니다.

## 덮이지 않은 것을 이름으로 적는다

- **B1의 두 갈래를 한 테스트가 본다.** `TestEnqueueAlertRequiresAKeyAndAType`는
  키 없음과 타입 없음을 둘 다 `t.Error`로 검사하지만, **어느 쪽이 실패했는지는
  `alertKey` 안의 분기**다. 그 분기는 이 번들의 것이 아니다.
- **`recordAlertTx`의 내부 분기는 이 표에 없다.** `EnqueueAlert`의 AST는 그 호출을
  한 줄로 본다. 재무장·dedup·`owed` 판정의 분기 지도는 `recordAlertTx` 쪽이고,
  a099는 그 함수의 제어 흐름을 안 바꿨다 — 바꾼 것은 재무장 UPDATE의 SET 절이다
  (`alertClaimCleared` 추가).
