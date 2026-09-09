# Branch Test Map: `TestFailedAttemptsAccumulateAndTheRowStaysPending`

`ast.json`의 열거가 정본이다: 분기 9 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:80` 배치 오류 가드 — 단언이 아니다 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B2 | `:87` 세 번 실패시킨다 — 한 claim 아래서 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B3 | `:90` 배치 오류 가드 — 단언이 아니다 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B4 | `:93` **매 시도의 정산이 `SettleApplied`다** — 임차가 예산 내내 유지된다 (a099가 더한 단언) | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B5 | `:100` 배치 오류 가드 — 단언이 아니다 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B6 | `:103` **상태가 PENDING이다** — critical 알림은 보존된다 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B7 | `:106` **`attempts`가 3이다** | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B8 | `:109` **`last_attempt_at`이 기록됐다** | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |
| B9 | `:112` **`last_error`가 마지막 이유다** | `TestFailedAttemptsAccumulateAndTheRowStaysPending` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입). 그리고 **§4.9 되돌림 관측**: `MarkAlertAttemptFailed`가 임차를 풀면 두 번째 반복이 `SettleLeaseLost`로 실패한다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **예산이 다 떨어졌을 때의 해제를 안 본다.** 이 테스트는 세 번 기록만 하고 `deliver`의 루프를 안 탄다.
- **시도 사이에 다른 발송자가 끼어드는 것을 안 본다.** `TestClaimingAnUndeliveredAlertStillOwesDelivery` `a096_claim_for_delivery_test.go:197`이 그 자리를 본다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
