# Branch Test Map: `Journal.ClaimAlertForDelivery`

**GREEN 칸은 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: 분기 8 · 이탈 9.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:235` `alertKey`가 거절한다 (키/타입 공백) | **이 진입점에는 없다** — 같은 `alertKey`를 `TestEnqueueAlertRequiresAKeyAndAType` `outbox_test.go:62`가 다른 진입점에서 본다 | no | **no (이 경로에는)** |
| B2 | `:238` 청구자 이름이 공백이다 | `TestClaimingWithoutANameIsRefused` `outbox_test.go:240` | **yes — 2026-08-12 관측.** 가드를 지우면 `an unnamed claimant took the lease` · `a blank claimant took the lease` · `rows = 1, want 0` 셋이 동시에 뜬다 | **yes** |
| B3 | `:242` `BeginTx` 실패 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| B4 | `:248` `recordAlertTx` 실패 | 없음 | no | **no** |
| B5 | `:251` **`!owed` — 보낼 것이 없다** | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` `a096_claim_for_delivery_test.go:64` · `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` `a096_claim_for_delivery_test.go:150` | **yes — §4.3 되돌림 관측**: `ClaimSettled` 대신 청구를 시도하면 정산된 행에 임차가 남는다 | **yes** |
| B6 | `:252` `!owed` 경로의 commit 실패 | 없음 | no | **no** |
| B7 | `:260` `acquireAlertClaimTx` 실패 | 없음 — 드라이버 오류 경로다. **「못 잡았다」는 이 분기가 아니다** | no | **no** |
| B8 | `:263` 청구 경로의 commit 실패 | 없음 | no | **no** |

## 정상 이탈 `:266` — a099가 존재하는 이유

| Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|
| **두 발송자가 한 PENDING 행에 닿으면 발송권은 하나** | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` `a099_claim_excludes_the_second_sender_test.go:41` | **yes — 2026-08-11 관측**: `senders granted the right to send = 2, want 1`. 예상대로 **둘 다 발송권을 받았다** | **yes** |
| 갓 기록된 행은 곧바로 잡힌다 | `TestClaimingAFreshAlertOwesDelivery` `a096_claim_for_delivery_test.go:37` | no (기존 통과) | **yes** |
| 만료된 임차는 다음 발송자가 가져간다 | `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70` | yes | **yes** |
| 짧은 임차를 든 자가 살아 있는 임차를 못 훔친다 | `TestALeaseHolderWithAShorterLeaseCannotStealALiveClaim` `a099_lease_lifecycle_test.go:178` | yes | **yes** |

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다. **예외는 B2다** — 그 테스트와 그 관측은
§5.6에서 이 표를 쓰다가 공백을 발견해 그 자리에서 만들었고, 가드 제거 → 실패 →
복원 순서로 관측했다(2026-08-12).

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B3 · B4 · B6 · B7 · B8** — 드라이버·커밋·위임 실패 경로다. a099는 이 다섯의
  조건도 반환도 새로 만들지 않았다(B7·B8은 청구 절반이 생기며 늘었지만 형태는
  기존 commit 오류 처리와 같다). **`not-applicable`: 이 change는 다섯을 근거로
  아무것도 주장하지 않는다.**
- **B1** — `alertKey`의 거절을 이 진입점으로 관측하는 테스트가 없다.
  **a099가 만든 공백이 아니다**: proposal 시점에도 키/타입 검증은
  `TestEnqueueAlertRequiresAKeyAndAType` 하나가 다른 진입점에서 봤다.

## 덮이지 않은 것을 이름으로 적는다

- **B7이 「임차를 못 잡았다」를 뜻하지 않는다.** 그것은 오류가 아니라
  `ClaimHeldElsewhere`이고 정상 이탈 `:266`으로 나간다. **2·3판 tasks가 이것을
  `owed=false`라고 적었고, 2라운드 A-P1이 깬 자리가 거기다.**
- **`recordAlertTx`의 분기는 이 표에 없다.** `internal-journal--recordalerttx`
  번들이 진다. 이 함수의 AST는 그 호출을 한 줄로 본다.
- **`acquireAlertClaimTx`의 판정 분기도 이 표에 없다.**
  `internal-journal--acquirealertclaimtx` 번들이 진다.
- **`-race` 미실행.** 두 트랜잭션의 배타성은 §5.2가 관측한다. 이 문서를 쓰는
  시점에 아직 안 돌았고, **돌기 전에 성립한다고 적지 않는다.**
