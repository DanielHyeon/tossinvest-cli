# Branch Test Map: `TestMarkingANonPendingAlertIsRefused`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:179` 배치 오류 가드 — 단언이 아니다 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B2 | `:182` 배치 오류 가드 — 단언이 아니다 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B3 | `:184` 첫 정산이 `SettleApplied`다 — 배치 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B4 | `:184` 같은 조건의 `else if` 짝 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B5 | `:189` 배치 오류 가드 — 단언이 아니다 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B6 | `:192` **두 번째 정산이 `SettleAlreadySettled`다** — 옛 `ErrAlertNotFound`의 절반 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B7 | `:197` 배치 오류 가드 — 단언이 아니다 | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |
| B8 | `:200` **없는 id는 `SettleNotFound`다** — 나머지 절반 (a099가 가른 것) | `TestMarkingANonPendingAlertIsRefused` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입). 그리고 **§4.4 되돌림 관측**: 0행일 때 같은 트랜잭션에서 다시 읽지 않으면 두 경우를 못 가른다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **`SettleLeaseLost`를 이 테스트가 안 본다.** 네 결과 중 셋만 본다 — 임차 상실은 `TestTheSameSenderNameCannotSettleAnEarlierLease` `a099_lease_lifecycle_test.go:266`가 본다.
- **토큰이 맞는데 상태가 틀린 경우와 상태가 맞는데 토큰이 틀린 경우를 따로 안 본다.** 여기는 앞의 것만이다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
