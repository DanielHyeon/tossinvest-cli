# Branch Test Map: `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`

`ast.json`의 열거가 정본이다: 분기 8 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:50` 배치 오류 가드 — 단언이 아니다 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B2 | `:53` 배치된 행이 PENDING이고 id가 1이다 — 경합 전 상태 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B3 | `:65` 발송자 여럿을 동시에 띄운다 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B4 | `:73` 각 결과를 세 갈래로 분류하는 `switch` | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B5 | `:74` 오류는 따로 모은다 — 배제가 아니라 고장이다 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B6 | `:76` **`ClaimAcquired`를 센다** — 발송권의 수 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B7 | `:84` 모은 오류를 보고한다 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |
| B8 | `:87` **발송권이 정확히 하나다** — a099의 전부 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 2026-08-11 관측**: `senders granted the right to send = 2, want 1`. 예상대로 **둘 다 발송권을 받았다.**

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **`-race`로 안 돌렸다.** 이 테스트가 관측하는 배타성은 SQLite의 잠금 모형에 기댄다. §5.2가 관측하고, **관측 전에는 성립한다고 적지 않는다.**
- **발송자 수를 하나만 쓴다.** 셋 이상으로 늘리면 다른 것이 보일 수 있지만 안 봤다.
- **어느 발송자가 이기는지는 안 본다** — 정해져 있지 않고 정해질 필요도 없다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**
