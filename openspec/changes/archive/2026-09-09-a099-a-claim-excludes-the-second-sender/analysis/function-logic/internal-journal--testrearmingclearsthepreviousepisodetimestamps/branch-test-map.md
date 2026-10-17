# Branch Test Map: `TestReArmingClearsThePreviousEpisodeTimestamps`

`ast.json`의 열거가 정본이다: 분기 12 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:156` 배치 오류 가드 — 단언이 아니다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B2 | `:160` 배치 오류 가드 — 단언이 아니다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B3 | `:163` 배치 오류 가드 — 단언이 아니다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B4 | `:165` **배치가 성립했다** — 두 도장이 찍혀 있다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B5 | `:165` 같은 조건의 `else if` 짝 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B6 | `:172` 두 번째 episode를 **다른 이유로** 청구한다 — 오류 가드가 여러 줄에 걸쳐 있다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B7 | `:175` **재무장 뒤 `ClaimAcquired`다** | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B8 | `:175` 같은 조건의 `else if` 짝 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B9 | `:180` 배치 오류 가드 — 단언이 아니다 | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B10 | `:183` **상태가 PENDING이다** | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B11 | `:186` **`delivered_at`이 nil이다** | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |
| B12 | `:191` **`last_attempt_at`이 nil이다** | `TestReArmingClearsThePreviousEpisodeTimestamps` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패.** §4.3 전에는 `ClaimAcquired`가 없다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **임차 열 넷은 이 표에 없다.** 같은 원칙이 같은 UPDATE에서 그것들도 지우지만, 그것을 보는 것은 `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340`다. **원칙은 하나인데 테스트가 둘이다.**
- **`created_at`은 안 지운다** — 조건이 처음 생긴 시각이라 episode에 속하지 않는다. 이 테스트가 그것을 **확인하지도 않는다.**
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
