# Branch Test Map: `TestAnUnrecognisedAlertStateOwesDelivery`

`ast.json`의 열거가 정본이다: 분기 6 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:252` 배치 오류 가드 — 단언이 아니다 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |
| B2 | `:255` 테스트가 직접 SQL로 알 수 없는 상태를 심는다 — 배치 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |
| B3 | `:261` 배치 오류 가드 — 단언이 아니다 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |
| B4 | `:264` **알 수 없는 상태는 `ClaimAcquired`다** — 열린 쪽 실패 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |
| B5 | `:268` **상태가 PENDING으로 복구됐다** — 다음 관측이 정상 경로를 걷는다 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |
| B6 | `:271` 복구된 행이 새 토큰으로 정산된다 — 임차가 복구와 함께 갱신됐다 | `TestAnUnrecognisedAlertStateOwesDelivery` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입). 그리고 **§4.6 되돌림 관측**: 복구 UPDATE가 임차를 안 지우면 B6이 실패한다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **어떤 알 수 없는 값인지는 하나만 본다.** 값의 종류에 따라 갈리는 것은 없다 — `claimOwed`의 `default`가 전부를 같이 다룬다.
- **`remindAfter`가 0일 때 이 복구가 여전히 도는지**는 여기가 아니라 `TestRecoveringAnUnknownStateIgnoresTheReminderWindow` `a097_rearm_is_a_new_episode_test.go:214`가 본다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
