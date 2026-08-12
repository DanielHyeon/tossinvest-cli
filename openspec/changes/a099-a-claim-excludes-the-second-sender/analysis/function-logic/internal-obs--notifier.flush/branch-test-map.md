# Branch Test Map: `Notifier.Flush`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 19 · 이탈 5 · defer 1.

**§5.6 갱신**: proposal 시점 6에서 11로 늘었다. 는 다섯은 전부 §4.11이
행마다 임차를 걸면서 생긴 것이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:728` `n.Journal == nil` | 기존 | no | **yes (기존)** |
| B2 | `:738` `PendingAlerts` 실패 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| B3 | `:741` `range pending` | 기존 | no | **yes (기존)** |
| B4 | `:742` `n.Publisher == nil` | 기존 | no | **yes (기존)** |
| B5 | `:753` **`ClaimAlertByID` 오류 — 이제 백로그를 안 버린다 (4라운드)** | 없음 | no | **no — 고침은 넣었고 핀이 없다** |
| B6 | `:759` 그 오류를 로그로 (4라운드 신설) | 없음 | no | **no** |
| B7 | `:764` `!= ClaimAcquired` — 건너뛴다 | `TestContentionNeitherLocksNorUnlocksTheEntryGate` | **yes — §4.11 되돌림 관측** | **yes** |
| B8 | `:767` 경합을 로그로 (`engine.alert_claim_held`, 4라운드에 `Warn` 으로) | `TestContentionLossAndTakeoverAreThreeEvents` | yes | **yes** |
| B9 | `:779` `Publish` 실패 | `TestFlushHandsTheLeaseBackWhenThePublishFails` `a099_round4_test.go` | **yes (4라운드)** | **yes** |
| B10 | `:782` `MarkAlertAttemptFailed` 자체가 오류 (4라운드 신설) | 없음 | no | **no** |
| B11 | `:786` **정산이 `SettleApplied` 가 아니다 (4라운드 신설)** | 없음 | no | **no** |
| B12 | `:783` 그 오류를 로그로 (4라운드 신설) | 없음 | no | **no** |
| B13 | `:786` (B11 의 else-if 짝) | 위와 같다 | no | **no** |
| B14 | `:803` **반납이 오류 (4라운드 신설)** | 없음 | no | **no** |
| B15 | `:807` **반납 결과가 `SettleApplied` 가 아니다 (4라운드 신설)** | 없음 | no | **no** |
| B16 | `:804` 그 오류를 로그로 (4라운드 신설) | 없음 | no | **no** |
| B17 | `:807` (B15 의 else-if 짝) | 위와 같다 | no | **no** |
| B18 | `:813` `MarkAlertDelivered` 실패 | 없음 | no | **no (기존부터 없다)** |
| B19 | `:816` 정산이 `SettleApplied` 가 아니다 — 남의 행 | 없음 | no | **no — 안 덮였다** |

이탈 `:653`(정상)은 분기가 아니다. `PendingAlerts` → 발송 → `UndeliveredCount`.

## 이 함수의 배제가 무엇을 사는가

| 시나리오 | a099 이전 | 지금 |
|---|---|---|
| flush와 관측이 **같은 `Notifier`** | `n.mu`가 가른다 | 그대로 |
| flush와 관측이 **다른 `Notifier`** | **아무것도 안 가른다** | **원장의 임차가 가른다** (B6) |
| 목록을 읽은 뒤 행이 정산됐다 | 그대로 또 보낸다 | **B6이 건너뛴다** |

**두 번째 줄이 D7의 근거다.** 「임차가 있는데 우회 경로가 하나 남으면 그것은
임차가 아니다」의 그 우회 경로가 이 함수였다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B1 · B3 · B4** — 배선 검사와 루프. a099가 안 건드린다.
- **B2 · B10** — 드라이버 오류 경로. a099 이전부터 안 덮여 있다.
  **`not-applicable`: 이 change는 둘을 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **B5 · B11이 a099가 만든 경로인데 안 덮였다.**
  - **B5**: 한 행의 청구가 드라이버 오류를 내면 **뒤의 행들을 안 돈다.**
  - **B11**: 발송은 나갔는데 정산이 남의 것이다. `logLeaseLost`만 하고 넘어간다.
  **§6.5 리뷰가 이것을 봐야 한다.**
- **B9의 해제(`:637`)를 직접 단언하는 테스트가 없다.** 같은 계약을 `deliver` 쪽
  `TestASenderThatSpendsItsBudgetReleasesTheRow`가 보지만, **이 루프의 해제는
  별도 코드이고 별도로 안 관측됐다.** 이름을 적어 둔다.
- **`:632`·`:637`이 반환을 버린다.** 임차 상실이 여기서 조용하다 —
  `deliver`가 같은 자리에서 `logLeaseLost`를 부르는 것과 대조된다.
- **프로덕션 호출자가 0이다.** 이 표의 GREEN은 전부 테스트 안의 관측이고,
  **운영에서 이 경로가 도는 것을 본 적이 없다.** a098이 호출자를 만든다.
