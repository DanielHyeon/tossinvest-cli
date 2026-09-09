# Branch Test Map: `TestClaimingAnUndeliveredAlertStillOwesDelivery`

`ast.json`의 열거가 정본이다: 분기 11 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:204` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B2 | `:208` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B3 | `:211` 실패 기록이 `SettleApplied`다 — **임차를 유지한 채로** 기록됐다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B4 | `:214` 행이 여전히 PENDING이다 — 실패는 배달이 아니다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B5 | `:219` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B6 | `:222` **두 번째 청구가 `ClaimSettled`가 아니다** — 빚은 그대로다 (옛 `owed=true`의 절반) | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B7 | `:225` **두 번째 청구가 `ClaimHeldElsewhere`다** — 남이 쥐고 있다 (a099가 가른 나머지 절반) | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B8 | `:232` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B9 | `:235` 해제 뒤 다시 청구하면 잡힌다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B10 | `:237` **해제 뒤 `ClaimAcquired`다** — 임차를 놓으면 빚이 다시 잡힌다 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |
| B11 | `:237` 같은 조건의 `else if` 짝 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입). 그리고 **§4.2·4.3 되돌림 관측**: 임차가 없으면 B7이 `ClaimAcquired`를 받아 실패한다 — 즉 **두 번째 발송자가 보낸다.**

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **만료로 풀리는 경로를 안 본다.** 여기서는 `ReleaseAlertClaim`으로 명시 해제한다. 만료는 `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70`다.
- **두 발송자가 정말 동시에 들어오는 것을 안 본다** — 순차다. 동시성은 `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` `a099_claim_excludes_the_second_sender_test.go:41`가 진다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
