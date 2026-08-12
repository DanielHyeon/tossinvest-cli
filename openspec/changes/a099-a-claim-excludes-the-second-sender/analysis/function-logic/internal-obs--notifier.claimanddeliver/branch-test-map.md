# Branch Test Map: `Notifier.claimAndDeliver`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 9 · 이탈 5 · defer 1.

**§5.6 갱신.** proposal 시점 분기는 4였다. 3판까지 이 줄은
*"제어 흐름을 안 바꾼다 — doc comment만 고친다"*였고 **계획과 반대였다.**
구현 후 분기는 **9**다: 처분 셋을 가르는 `switch`(B4~B6), 경합 로그(B7),
탈취 로그(B8), 임차 상실(B9).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:252` claim이 오류를 준다 — 원장에 못 썼다 | 기존 (a097) | no | **yes (기존)** |
| B2 | `:265` `n.Log`가 배선돼 있다 | 기존 | no | **yes (기존)** |
| B3 | `:268` `n.Gate`가 배선돼 있다 — 진입 래치 | 기존 (a097) | no | **yes (기존)** |
| B4 | `:273` 처분 셋으로 갈리는 `switch` | `TestContentionLossAndTakeoverAreThreeEvents` `a099_lease_events_test.go:162` | yes | **yes** |
| B5 | `:274` **`ClaimSettled` — 보낼 것이 없다** | 기존 (a096 reminder window) | no | **yes (기존)** |
| B6 | `:283` **`ClaimHeldElsewhere` — 다른 발송자가 쥐고 있다** | `TestContentionLossAndTakeoverAreThreeEvents` `a099_lease_events_test.go:162` · `TestContentionNeitherLocksNorUnlocksTheEntryGate` `a099_regression_pins_test.go:120` | **yes — 2026-08-11**: 구현 전 `publishes while one sender held the row = 2, want 1` | **yes** |
| B7 | `:293` 경합을 로그로 남긴다 (`engine.alert_claim_held`) | `TestContentionLossAndTakeoverAreThreeEvents` `a099_lease_events_test.go:162` | **yes — §4.5 되돌림 관측**: 로그를 빼면 세 사건 중 하나가 안 뜬다 | **yes** |
| B8 | `:303` **만료된 임차를 뺏었다** (`engine.alert_claim_stolen`) | `TestContentionLossAndTakeoverAreThreeEvents` `a099_lease_events_test.go:162` | **yes — §4.5 되돌림 관측** | **yes** |
| B9 | `:312` **publish 도중 임차를 잃었다** | `TestASenderThatLosesTheLeaseStopsAtOnce` `a099_lease_events_test.go:135` | **yes — §4.9 되돌림 관측**: `lost`를 무시하면 임차를 잃은 발송자가 `owed=true`를 돌려주고 호출자가 격상한다 | **yes** |

## 네 이탈이 무엇을 뜻하는가

| 이탈 | `sent, owed, err` | 뜻 |
|---|---|---|
| `:271` | `false, false, err` | **원장에 못 썼다.** 진입 게이트가 잠긴다 |
| `:282` | `false, false, nil` | 정산됨 — 보낼 것이 없다 |
| `:300` | `false, false, nil` | **경합에서 졌다 — 게이트를 양방향으로 안 건드린다** |
| `:316` | `sent, false, nil` | **보내는 중에 임차를 잃었다** — 격상하지 않는다 |
| `:318` | `sent, true, nil` | 정상 |

**`:300`과 `:316`이 `owed=false`인 것이 a099의 핵심 계약이다.**
경합에서 지는 것은 배달 실패가 아니다. 호출자(`Notify`)가 격상하면
**성공한 배달 뒤에 사람만 열 수 있는 잠금이 남는다.**

## R3이 관측하는 것은 이 함수의 분기가 아니라 **그 밖**이다

`n.mu`(`:248`)는 **한 `Notifier` 인스턴스**의 것이다.

| 발송자 둘이 | 배제 |
|---|---|
| 같은 `Notifier`를 공유한다 | `n.mu`가 가른다 |
| **다른 `Notifier`거나 원장을 직접 쓴다** | **원장의 임차만 가른다** |

`TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`
(`a099_claim_excludes_the_second_sender_test.go:41`)가 두 번째 줄을 관측한다 —
`Notifier`를 거치지 않고 원장에 직접 두 claim을 건다.
obs 쪽의 대응은 `a099Pair`가 한 원장 위에 `Notifier` **둘**을 세워
뮤텍스를 가르는 방식이다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B1 · B2 · B3 · B5** — a096·a097이 만든 경로이고 기존 테스트가 덮는다.
  a099가 조건도 반환도 안 바꿨다.
  **`not-applicable`: 이 change는 이 넷을 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **`ClaimAlertForDelivery` 호출(`:251`)이 더한 UPDATE 하나의 지연.**
  이 함수의 분기가 아니라 호출 비용이다. **§5.7이 재고, 재기 전에는
  불변식 4를 만족한다고 적지 않는다.**
- **`n.mu`의 구간** — a099의 대상이 아니다. 줄이는 것은 a092의 D0.3a·D0.3b이고,
  a099는 그 재설계의 **전제만** 만들었다.
- **B8의 `n.Log == nil` 쪽이 안 덮였다.** 로그 없이 도는 `Notifier`에서
  탈취가 조용히 지나가는 경로다. 조건이 `claim.Stole && n.Log != nil`이라
  **로그가 없으면 탈취 사실이 아무 데도 안 남는다.** 프로덕션은 항상 로그를
  배선하므로 실질 위험은 낮지만, **이름은 적어 둔다.**
