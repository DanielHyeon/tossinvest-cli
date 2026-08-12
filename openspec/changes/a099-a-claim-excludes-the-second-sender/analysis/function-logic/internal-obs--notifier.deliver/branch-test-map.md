# Branch Test Map: `Notifier.deliver`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 24 · 이탈 6 · defer 0.

**§5.6 갱신**: proposal 시점 12에서 24로 늘었다. `else` 절과 그 안의 `if`를
AST가 따로 세므로 **판정의 수는 24보다 적다** — B13/B15, B14/B16, B19/B21,
B20/B22가 두 개씩 짝이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:424` `attempts <= 0`이면 기본값 3 | 기존 | no | **yes (기존)** |
| B2 | `:430` 재시도 루프 | `TestFailedAttemptsAccumulateAndTheRowStaysPending` `outbox_test.go:74` (원장 쪽) | no | **yes (기존)** |
| B3 | `:431` `n.Publisher == nil` | 기존 | no | **yes (기존)** |
| B4 | `:436` publish 성공 | 기존 | no | **yes (기존)** |
| B5 | `:438` 정산 호출이 오류를 안 냈다 | 기존 | no | **yes (기존)** |
| B6 | `:439` **정산 결과 `switch`** | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` | **yes — §4.8 되돌림 관측** | **yes** |
| B7 | `:440` `SettleApplied` — 정상 | 기존 | no | **yes (기존)** |
| B8 | `:442` **`SettleLeaseLost`/`SettleAlreadySettled` — 보냈지만 남의 행** | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` | **yes — §4.8** | **yes** |
| B9 | `:454` `SettleNotFound` — 행이 사라졌다 | **없음** | no | **no — 안 덮였다** |
| B10 | `:458` `markErr == nil` | 기존 | no | **yes (기존)** |
| B11 | `:474` `n.Log != nil` (정산 실패) | 기존 (a096 r2) | no | **yes (기존)** |
| B12 | `:479` `n.Gate != nil` (같음) — **게이트 래치** | `TestAPublishedButUnsettledRowKeepsItsLease` `a099_regression_pins_test.go:48` | **R8은 planned RED가 아니다 — 회귀 핀이다** | **yes** |
| B13 | `:491` `MarkAlertAttemptFailed`가 오류 | 없음 | no | **no** |
| B15 | `:492` `n.Log != nil` (같음) | 없음 | no | **no** |
| B14 | `:495` B13의 `else` | `TestRecordingAFailedAttemptKeepsTheLease` `a099_regression_pins_test.go:28` (원장 쪽) | — | **yes** |
| B16 | `:495` **`failed.Outcome != SettleApplied` — 임차 상실** | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` | **yes — §4.8 되돌림 관측**: 무시하면 임차를 잃은 발송자가 예산을 계속 쓴다 | **yes** |
| B17 | `:502` `attempt < attempts` | 기존 | no | **yes (기존)** |
| B18 | `:503` `!n.wait(ctx)` | 기존 | no | **yes (기존)** |
| B19 | `:517` `ReleaseAlertClaim`이 오류 | **없음** | no | **no — 안 덮였다** |
| B21 | `:518` `n.Log != nil` (같음) | 없음 | no | **no** |
| B20 | `:521` B19의 `else` | `TestASenderThatSpendsItsBudgetReleasesTheRow` `a099_lease_events_test.go:61` | **yes — §4.10 되돌림 관측** | **yes** |
| B22 | `:521` 해제 시점에 이미 임차를 잃었다 | **없음** | no | **no — 안 덮였다** |
| B23 | `:526` `n.Log != nil` (예산 소진) | 기존 | no | **yes (기존)** |
| B24 | `:531` `n.Gate != nil` (같음) — **게이트 래치** | 기존 | no | **yes (기존)** |

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
