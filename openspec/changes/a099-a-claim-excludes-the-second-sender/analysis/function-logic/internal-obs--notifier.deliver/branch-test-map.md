# Branch Test Map: `Notifier.deliver`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 27 · 이탈 6 · defer 0.

**§5.6 갱신**: proposal 시점 12에서 24로 늘었다. `else` 절과 그 안의 `if`를
AST가 따로 세므로 **판정의 수는 24보다 적다** — B13/B15, B14/B16, B19/B21,
B20/B22가 두 개씩 짝이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:422` `attempts <= 0` | 기존 | no | **yes (기존)** |
| B2 | `:428` 재시도 루프 | 기존 | no | **yes (기존)** |
| B3 | `:429` `Publisher == nil` | 기존 | no | **yes (기존)** |
| B4 | `:434` 발행 성공 | 기존 | no | **yes (기존)** |
| B5 | `:436` `MarkAlertDelivered` 가 오류를 안 냈다 | 기존 | no | **yes** |
| B6 | `:437` `switch settled.Outcome` | — (아래 네 arm) | — | — |
| B7 | `:438` `SettleApplied` — 배달 성립 | 기존 | no | **yes (기존)** |
| B8 | `:440` `SettleLeaseLost`·`SettleAlreadySettled` — 나갔지만 내 행이 아니다 | `TestASenderThatLosesTheLeaseStopsAtOnce` | yes | **yes** |
| B9 | `:452` `SettleNotFound` — 원장이 행을 잃었다 | `TestARowThatVanishedIsNotReportedAsContention` `a099_round4_test.go` | **yes (4라운드)** | **yes** |
| B10 | `:454` **`default` — 이 빌드가 모르는 outcome (4라운드 신설)** | 없음 — 다섯째 값이 아직 없다 | no | **no — 도달 불가, 방향만 fail-safe 로 뒤집었다** |
| B11 | `:478` 정산 실패를 로그로 | 기존 | no | **yes (기존)** |
| B12 | `:483` 정산 실패가 게이트를 잠근다 | 기존 a096 라운드2 | no | **yes (기존)** |
| B13 | `:495` `MarkAlertAttemptFailed` 자체가 오류 | 없음 — 원장 쓰기 실패 주입 seam 이 없다 | no | **no — `not-applicable`, review §11.6** |
| B14 | `:499` `failed.Outcome != SettleApplied` | `TestASenderThatLosesTheLeaseStopsAtOnce` | yes | **yes** |
| B15 | `:496` 그 오류를 로그로 | 없음 (B13 과 같은 이유) | no | **no** |
| B16 | `:499` (B14 의 else-if 짝) | 위와 같다 | yes | **yes** |
| B17 | `:509` **전송 실패를 먼저 남긴다 (4라운드 신설)** | `TestARowThatVanishedIsNotReportedAsContention` | **yes** | **yes** |
| B18 | `:519` **`SettleNotFound` 는 게이트를 잠근다 (4라운드 신설)** | `TestARowThatVanishedIsNotReportedAsContention` | **yes** | **yes** |
| B19 | `:525` 마지막 시도가 아니다 | 기존 | no | **yes (기존)** |
| B20 | `:526` `wait` 가 false — **컨텍스트가 끝났다** | `TestACancelledSenderStillHandsTheLeaseBack` `a099_round4_test.go` | **yes (4라운드)** | **yes** |
| B21 | `:543` `switch` — 반납 결과 (4라운드 신설) | — (아래 세 arm) | — | — |
| B22 | `:544` 반납 자체가 오류 | 없음 | no | **no** |
| B23 | `:545` 그 오류를 로그로 | 없음 | no | **no** |
| B24 | `:548` `SettleApplied` — 정상 반납, 게이트로 간다 | `TestASenderThatSpendsItsBudgetReleasesTheRow` · `TestACancelledSenderStillHandsTheLeaseBack` | yes | **yes** |
| B25 | `:551` **`default` — 반납 때 행이 남의 것이었다 (4라운드 신설)** | 없음 — `markErr != nil` 창을 테스트로 못 연다 | no | **no — `not-applicable`, review §11.6** |
| B26 | `:565` 예산 소진을 로그로 | 기존 | no | **yes (기존)** |
| B27 | `:570` 예산 소진이 게이트를 잠근다 | 기존 | no | **yes (기존)** |

## 여섯 이탈이 임차를 어떻게 남기는가

| 이탈 | `sent, lost` | 임차 | Test |
|---|---|---|---|
| `:441` | `true, false` | **푼다** (정산 UPDATE가 같이) | 기존 |
| `:453` | `true, true` | 남의 것 — 안 건드린다 | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` |
| `:459` | `true, false` | (도달 불가에 가깝다 — B9만 여기로 온다) | 없음 |
| `:487` | `false, false` | **유지한다 — 만료가 푼다** | `TestAPublishedButUnsettledRowKeepsItsLease` `a099_regression_pins_test.go:48` |
| `:500` | `false, true` | 남의 것 — 즉시 중단 | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` |
| `:534` | `false, false` | **`ReleaseAlertClaim`으로 푼다** | `TestASenderThatSpendsItsBudgetReleasesTheRow` `a099_lease_events_test.go:61` |

**`:487`이 이 표에서 가장 중요한 줄이다.** 여기서 임차를 풀면 a096 폭풍이
성공 경로로 돌아온다. R8이 그것을 고정한다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B1 · B3 · B4 · B5 · B7 · B10 · B11 · B17 · B18 · B23 · B24** — a096·a097이 만든
  경로이고 기존 테스트가 덮는다. a099가 조건도 반환도 안 바꿨다.
  **`not-applicable`: 이 change는 이들을 근거로 아무것도 주장하지 않는다.**
- **B13 · B15** — `MarkAlertAttemptFailed`의 **드라이버 오류** 경로.
  a099 이전부터 로그만 남기고 루프를 계속 돈다. 안 바꿨다.

## 덮이지 않은 것을 이름으로 적는다

- **B9 · B19 · B22가 a099가 만든 경로인데 안 덮였다.**
  - **B9** `SettleNotFound`: 배달 도중 행이 사라졌다.
  - **B19** 해제 자체가 드라이버 오류를 낸다.
  - **B22** 예산을 다 쓰고 해제하려는 순간 이미 임차가 남의 것이다.
  셋 다 `SettleOutcome`이 넷으로 갈리면서 생긴 자리이고, **셋 다 관측되지 않았다.**
  **§6.5 리뷰가 이것을 봐야 한다.**
- **이탈 `:459`는 「인식 못 하는 `SettleOutcome`」 전용 통로다.** AST 열거로
  경로를 다 따라가면 그것만 남는다:
  `SettleApplied`는 `:441`로, `SettleLeaseLost`/`SettleAlreadySettled`는 `:453`으로
  이미 나갔고, `SettleNotFound`(B9)는 `markErr`를 채워 B10을 거짓으로 만든다.
  B5가 거짓이면 애초에 `markErr != nil`이다. **네 값 중 어느 것도 여기 못 온다.**
  `switch`에 `default`가 없으므로, **선언된 넷 밖의 값**만 `markErr == nil`인 채로
  `:458`을 통과한다.
  **그때 이 함수는 `true, false` — 「보냈고 정산됐다」를 돌려준다.**
  방향이 반대다: 모르는 결과는 「정산 안 됨」으로 다뤄야 게이트가 잠기고
  닫힌 쪽으로 실패한다. **오늘 원장이 그런 값을 안 만들므로 결함은 아니지만,
  a099가 만든 자리이고 §6.5 리뷰가 판정해야 한다.**
- **`Publish`의 실제 기한** — Publisher 구현이 정한다. §5.7이 실측한다.
