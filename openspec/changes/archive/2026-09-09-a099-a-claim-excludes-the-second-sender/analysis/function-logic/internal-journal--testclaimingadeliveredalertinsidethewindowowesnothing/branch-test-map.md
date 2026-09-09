# Branch Test Map: `TestClaimingADeliveredAlertInsideTheWindowOwesNothing`

`ast.json`의 열거가 정본이다: 분기 7 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:71` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B2 | `:74` 첫 청구가 `ClaimAcquired`다 — 배치가 성립했다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B3 | `:77` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B4 | `:84` 배치 오류 가드 — 단언이 아니다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B5 | `:87` 같은 조건은 같은 행이다 — id가 안 바뀐다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B6 | `:90` **창 안에서는 `ClaimSettled`다** — 옛 `owed=false`. 운영자는 방금 들었다 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |
| B7 | `:94` **억제된 청구가 행을 안 건드린다** — 상태가 `DELIVERED` 그대로 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패.** §4.3 전에는 `ClaimSettled`가 없고, §4.8 전에는 `MarkAlertDelivered`가 토큰을 안 받는다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **임차 열을 안 본다.** 정산된 행에 임차가 안 남는지는 `TestClaimingADeliveredAlertInsideTheWindowOwesNothing`이 아니라 `MarkAlertDelivered`의 UPDATE가 진다.
- **`B5`가 실패해도 `B6`가 돈다** (`t.Errorf`). 두 단언이 독립이라 출력이 둘 다 뜰 수 있다 — 의도된 것이다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
