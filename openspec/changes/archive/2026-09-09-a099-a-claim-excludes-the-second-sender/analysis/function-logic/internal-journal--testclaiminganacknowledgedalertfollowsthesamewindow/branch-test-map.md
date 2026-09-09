# Branch Test Map: `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:157` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B2 | `:164` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B3 | `:169` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B4 | `:171` **창 안에서는 `ClaimSettled`다** — 확인 뒤 재발생은 아직 뉴스가 아니다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B5 | `:171` 같은 조건의 `else if` 짝 — AST가 `else`와 `if`를 따로 센다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B6 | `:176` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B7 | `:178` **창이 지나면 `ClaimAcquired`다** — 확인 뒤의 재발생은 뉴스다 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |
| B8 | `:178` 같은 조건의 `else if` 짝 | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패.** §4.3 전에는 `ClaimSettled`·`ClaimAcquired`가 없다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **확인이 임차를 지우는지 안 지우는지를 이 테스트가 안 본다.** 그것은 `TestAcknowledgementIgnoresTheLease` `a099_regression_pins_test.go:101`다.
- **확인한 사람의 이름을 안 본다.** `outbox_test.go`의 `TestDeliveryAndAcknowledgementAreDistinctStates`가 본다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
