# Branch Test Map: `acquireAlertClaimTx`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 8.
**GREEN 칸은 실측해서 채운다.**

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:218` 행 읽기 실패 | 없음 | no | **no** |
| B2 | `:219` 없는 id → `ErrAlertNotFound` | 없음 — **도달 가능한데 안 덮였다** | no | **no** |
| B3 | `:226` 토큰 발급 실패 | 없음 — `crypto/rand` 실패 주입 없음 | no | **no** |
| B4 | `:237` UPDATE 실패 | 없음 | no | **no** |
| B5 | `:241` `RowsAffected` 실패 | 없음 | no | **no** |
| B6 | `:244` **잡았다 — 임차 열 넷을 쓴다** | `TestClaimingAFreshAlertOwesDelivery` `a096_claim_for_delivery_test.go:37` · `TestMigrationV30ToV31LeavesExistingAlertsClaimable` `migration_v31_test.go:18` | **yes — §4.2 되돌림 관측**: UPDATE를 빼면 토큰이 안 나온다 | **yes** |
| B7 | `:253` **만료된 남의 임차를 뺏었다** | `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70` · `TestContentionLossAndTakeoverAreThreeEvents` `a099_lease_events_test.go:162` | **yes — §4.4 되돌림 관측**: `Stole` 표시를 빼면 훔친 사실이 로그에 안 남는다 | **yes** |
| B8 | `:263` 0행 + PENDING이 아니다 → `ClaimSettled` | 없음 — **`ClaimAlertByID`로만 도달 가능하고 안 덮였다** | no | **no** |

## 정상 이탈 `:266` — `ClaimHeldElsewhere`

| Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|
| **두 발송자, 발송권은 하나** | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` `a099_claim_excludes_the_second_sender_test.go:41` | **yes — 2026-08-11**: `senders granted the right to send = 2, want 1` | **yes** |
| 살아 있는 임차 안에서는 두 번째가 못 잡는다 | `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70` (앞 절반) | yes | **yes** |
| 짧은 임차를 든 자가 못 훔친다 | `TestALeaseHolderWithAShorterLeaseCannotStealALiveClaim` `a099_lease_lifecycle_test.go:178` | **yes — R12** | **yes** |
| 발급 직후 skew 경계에서 안 뺏긴다 | `TestAClaimJustIssuedIsNotStolenAtTheSkewBoundary` `a099_lease_lifecycle_test.go:150` | yes | **yes** |
| 시계가 뒤로 갔으면 다시 열린다 | `TestAClaimIssuedInTheFutureIsReopened` `a099_lease_lifecycle_test.go:119` | yes | **yes** |

## 술어 네 갈래와 그것을 보는 테스트

| `alertClaimable` 갈래 | 뜻 | Test |
|---|---|---|
| `claim_token = ''` | 아무도 안 잡았다 | `TestClaimingAFreshAlertOwesDelivery` `a096_claim_for_delivery_test.go:37` |
| `claim_expires_at IS NULL` | v30에서 올라온 행 | `TestMigrationV30ToV31LeavesExistingAlertsClaimable` `migration_v31_test.go:18` |
| `claim_expires_at <= now` | 만료 — **포함(inclusive)** | `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70` |
| `claimed_at > now+skew` | 역행 시계 — **엄격 초과** | `TestAClaimIssuedInTheFutureIsReopened` `a099_lease_lifecycle_test.go:119` · 경계는 `TestAClaimJustIssuedIsNotStolenAtTheSkewBoundary` `a099_lease_lifecycle_test.go:150` |

**네 갈래에 각각 테스트가 있다.** 이것이 「술어를 한 자리에 뒀다」의 대가이자 이득이다 —
Go 분기가 아니라서 AST의 branches에는 안 잡히고, 그래서 **이 표가 따로 필요하다.**

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B1 · B3 · B4 · B5** — 드라이버·`crypto/rand` 오류 경로다.
  **`not-applicable`: 이 change는 넷을 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **B2와 B8은 도달 가능한데 안 덮였다.** 둘 다 `ClaimAlertByID` 경로다 —
  `Flush`가 `PendingAlerts`로 목록을 읽고 그 사이에 행이 사라지거나(B2)
  정산되면(B8) 온다. `Flush`는 B2를 `notifier.go:693`에서 이름으로 걸러 내고
  B8은 `ClaimHeldElsewhere`와 함께 건너뛴다. **경로는 살아 있고 테스트만 없다.**
  a099가 만든 공백이고, **§6.5 리뷰가 이것을 봐야 한다.**
- **`-race` 미실행.** 「같은 순간 둘이 들어오면 하나만 `n == 1`」은 SQLite의
  잠금 모형에 기댄다. §5.2가 관측하고, **관측 전에는 성립한다고 적지 않는다.**
