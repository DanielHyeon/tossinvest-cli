# Branch Test Map: `Journal.recordAlertTx`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 7.
**GREEN 칸은 실측해서 채운다.**

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:280` `alertKey` 거절 | **이 진입점에는 없다** — 두 호출자가 이미 통과시킨 뒤에만 도달한다 | no | **no (사실상 도달 불가)** |
| B2 | `:292` SELECT 결과로 갈린다 | `TestEnqueueAlertIsIdempotentOnTheEventKey` `outbox_test.go:30` (두 갈래 다) | no | **yes (기존)** |
| B3 | `:293` 기존 행을 찾았다 | `TestEnqueueAlertIsIdempotentOnTheEventKey` `outbox_test.go:30` | no | **yes (기존)** |
| B4 | `:295` **재무장한다** | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` `a096_claim_for_delivery_test.go:106` · `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340` | **yes — §4.6 되돌림 관측**: `alertClaimCleared`를 재무장 UPDATE에서 빼면 새 episode가 이전 임차를 물려받는다 | **yes** |
| B5 | `:334` 재무장 UPDATE 실패 | 없음 | no | **no (기존부터 없다)** |
| B6 | `:347` `!errors.Is(err, ErrNoRows)` | 없음 | no | **no** |
| B7 | `:355` INSERT 실패 | 없음 | no | **no** |
| B8 | `:359` `LastInsertId` 실패 | 없음 | no | **no** |

## 두 정상 이탈

| 이탈 | Scenario | Test | GREEN observed |
|---|---|---|---|
| `:346` | **기존 행 — 재무장 없이 `owed`만** | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` `a096_claim_for_delivery_test.go:64` · `TestARecordOnlyCallerDoesNotReArmADeliveredRow` `a097_rearm_is_a_new_episode_test.go:249` | **yes** |
| `:363` | 새 행 — `AlertPending`, `owed=true` | `TestClaimingAFreshAlertOwesDelivery` `a096_claim_for_delivery_test.go:37` · `TestARecordedAlertCanBeClaimedImmediately` `a099_lease_lifecycle_test.go:305` | **yes** |

## 이탈 `:346`이 a099의 반증이 나온 자리다

B4가 거짓이면(=PENDING 행) **UPDATE가 한 번도 안 돈다.** 행에 발송자 표시가
하나도 안 남고, 두 번째로 묻는 자도 같은 `owed=true`를 받는다.
`TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`가 구현 전 관측한
`senders granted the right to send = 2, want 1`이 이 경로다.

**a099가 그것을 고친 자리는 이 함수가 아니다.** 여기는 여전히 행을 안 건드린다.
배제는 `acquireAlertClaimTx`의 UPDATE 하나가 진다 —
`internal-journal--acquirealertclaimtx` 번들.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B5 · B6 · B7 · B8** — 전부 드라이버 오류 경로다.
  **`not-applicable`: 이 change는 넷을 근거로 아무것도 주장하지 않는다.**
- **B1** — 사실상 도달 불가능하다. 두 호출자가 트랜잭션을 열기 전에 같은
  `alertKey`를 이미 통과시킨다. 테스트로 도달시키려면 이 함수를 직접 부르는
  패키지 내부 테스트를 써야 하고, **그것은 도달 불가능한 분기를 도달 가능하게
  보이도록 만드는 일**이라 안 쓴다.

## 덮이지 않은 것을 이름으로 적는다

- **B4의 UPDATE는 열 열둘을 되돌리는데, 열두 개를 각각 확인하는 테스트는 없다.**
  a097의 네 테스트가 episode 관련 열들을, `TestReArmingClearsTheLeaseOfThePreviousEpisode`가
  임차 열 넷을 본다. **`payload`를 직접 읽는 테스트는 없다.**
  a099가 만든 공백이 아니다 — a097이 남긴 것이고 여기 적어 둔다.
- `claimOwed`의 판정 분기는 이 표에 없다. `internal-journal--claimowed` 번들이 진다.
