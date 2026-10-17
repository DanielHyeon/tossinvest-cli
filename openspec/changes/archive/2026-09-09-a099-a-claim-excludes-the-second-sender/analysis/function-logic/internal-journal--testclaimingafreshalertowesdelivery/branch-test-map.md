# Branch Test Map: `TestClaimingAFreshAlertOwesDelivery`

`ast.json`의 열거가 정본이다: 분기 4 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:46` 배치 오류 가드 — 단언이 아니다 | `TestClaimingAFreshAlertOwesDelivery` | (아래) | **yes** |
| B2 | `:49` 삽입된 행의 id가 0이 아니다 — 기록이 실제로 일어났다 | `TestClaimingAFreshAlertOwesDelivery` | (아래) | **yes** |
| B3 | `:52` **처분이 `ClaimAcquired`다** — 방금 삽입된 행은 보낸 적이 없다 (옛 `owed=true`) | `TestClaimingAFreshAlertOwesDelivery` | (아래) | **yes** |
| B4 | `:56` **토큰이 비어 있지 않다** — 없으면 이 발송을 정산할 수단이 없다 (a099가 더한 단언) | `TestClaimingAFreshAlertOwesDelivery` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패.** §4.3 전에는 `ClaimResult`도 `ClaimAcquired`도 없다. 그 시점에 이 단언은 쓸 수 없다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **토큰의 내용은 안 본다.** 비어 있지 않다는 것만 본다. 엔트로피·유일성은 `mintAlertClaimToken`의 몫이고 이 테스트가 안 진다.
- **두 번째 청구를 안 한다.** 이 테스트는 배제를 안 본다 — 그것은 `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
