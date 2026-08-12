# Branch Test Map: `TestAZeroReminderWindowNeverReArms`

`ast.json`의 열거가 정본이다: 분기 6 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:285` 배치 오류 가드 — 단언이 아니다 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |
| B2 | `:288` 배치 오류 가드 — 단언이 아니다 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |
| B3 | `:293` 배치 오류 가드 — 단언이 아니다 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |
| B4 | `:295` **창이 0이면 정산된 행은 `ClaimSettled`로 남는다** — 재무장이 없다 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |
| B5 | `:295` 같은 조건의 `else if` 짝 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |
| B6 | `:298` **행이 `DELIVERED` 그대로다** — 0이 상태를 안 건드린다 | `TestAZeroReminderWindowNeverReArms` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패.** §4.3 전에는 `ClaimSettled`가 없다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **음수 `remindAfter`를 안 본다.** `claimOwed`가 `<= 0`으로 같이 다루지만 이 테스트는 0만 준다.
- **0이 상태 복구까지 끄지 않는다**는 것은 여기서 안 본다 — `TestRecoveringAnUnknownStateIgnoresTheReminderWindow` `a097_rearm_is_a_new_episode_test.go:214`가 본다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
