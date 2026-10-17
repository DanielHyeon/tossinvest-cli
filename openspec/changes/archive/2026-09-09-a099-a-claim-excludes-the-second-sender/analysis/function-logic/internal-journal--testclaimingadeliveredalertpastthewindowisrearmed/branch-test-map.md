# Branch Test Map: `TestClaimingADeliveredAlertPastTheWindowIsReArmed`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:113` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B2 | `:116` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B3 | `:123` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B4 | `:126` 재알림은 같은 행이다 — 새 행이 아니다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B5 | `:129` **창이 지나면 `ClaimAcquired`다** — 재알림은 반드시 나간다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B6 | `:132` **상태가 PENDING으로 돌아왔다** — 첫 배달 경로를 걷는다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B7 | `:138` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |
| B8 | `:141` **재무장된 행의 정산이 `SettleApplied`다** — 새 episode의 임차로 정산된다 (a099가 더한 단언) | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입), 그리고 **§4.6 되돌림 관측**: 재무장 UPDATE에서 `alertClaimCleared`를 빼면 B8이 `SettleLeaseLost`로 실패한다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **이전 episode의 토큰으로 정산을 시도하지 않는다.** 그것을 보는 것은 `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340`다.
- **창 경계 정확히(`== claimRemind`)를 본다.** 경계 바로 앞은 앞 테스트가 본다. 그 둘 사이에 다른 값은 안 본다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
