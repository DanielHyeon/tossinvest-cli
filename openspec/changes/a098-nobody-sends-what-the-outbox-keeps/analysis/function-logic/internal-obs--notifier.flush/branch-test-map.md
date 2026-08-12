# Branch Test Map: `Notifier.Flush`

Source: `internal/obs/notifier.go` (427-462). AST 기준 branches 6 / returns 4.

**프로덕션 호출자는 0이다.** 아래 GREEN 칸의 테스트 셋은 전부 `Flush`를 **직접** 부르는
단위 테스트이고, 그것이 이 함수가 오늘 실행되는 유일한 방법이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:428` 원장이 nil → 조용히 0 | **없음** — `Flush`를 부르는 테스트 셋이 전부 원장을 배선한다 | no | **no** |
| B2 | `:438` `PendingAlerts` 오류 | **없음** — 배수 중에 원장을 깨는 테스트가 없다 | no | **no** |
| B3 | `:441` 밀린 행을 돈다 | `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` (`obs_test.go:427`) — PENDING 1행에서 `delivered=1`을 단언한다(`:443`) | no | yes |
| B4 | `:442` publisher가 nil → `break`, **기록 없음** | **없음** — 셋 다 publisher를 배선한다 | no | **no** |
| B5 | `:451` `Publish` 오류 → 시도 기록 후 `continue` | **없음** — 아래 참조 | no | **no** |
| B6 | `:455` `MarkAlertDelivered` 오류 → 배치를 끊는다 | **없음** | no | **no** |

## 여섯 중 다섯이 안 덮여 있다 — 그리고 그것이 a098의 위험 목록이다

`Flush`를 부르는 테스트는 트리 전체에 **셋뿐**이고 셋 다 **성공 경로만** 돈다.

| 테스트 | publisher | 어느 분기까지 가는가 |
|---|---|---|
| `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` (`obs_test.go:427`) | `failingPublisher{fail:true}` → **`setFail(false)` 후** flush | B3만 |
| `TestCriticalAlertSurvivesAProcessRestart` (`obs_test.go:551`) | `&failingPublisher{}` — `fail` 없음 | B3만 |
| `TestFlushCannotPublishBesideASend` (`a097_exclusion_is_an_event_test.go:92`) | `newReentrantPublisher()` — 발송은 성공 | B3만 |

**셋 다 B5를 안 탄다.** 첫 번째는 실패를 만들어 놓고 **flush 직전에 끈다**(`obs_test.go:438`
`pub.setFail(false)`) — 이름이 "recovered"인 이유이고, 그래서 flush 안의 실패 경로는
한 번도 실행되지 않는다. **테스트 이름이 실패를 말한다고 실패 분기를 타는 것은 아니다.**

## 필요한 RED — tasks §3의 R1~R5가 이 표에서 나온다

| # | Branch | Scenario | 기대 |
|---|---|---|---|
| R1 | B1 | 원장 없는 구성에서 루프가 돈다 | 루프가 **죽지 않고**, 아무것도 안 한다는 사실이 **한 번은 기록된다** |
| R2 | B2 | 배수 중 `PendingAlerts`가 오류 | 루프가 **죽지 않는다.** 오류는 로그 한 줄 |
| R3 | B4 | publisher가 nil인 채로 루프가 돈다 | **오늘 완전히 침묵한다** — 매 주기 `break`, 기록 0. 루프 쪽에서 들리게 한다 |
| R4 | B5 | `Publish`가 실패하는 채로 배수 | 행이 PENDING으로 남고 `attempts`가 는다 |
| R5 | B6 | `MarkAlertDelivered`가 실패 | `remaining`이 **0으로 보고된다**(`:456`) — 루프가 그것을 "다 비웠다"로 읽지 않는다 |

**R3이 a098이 지는 가장 중요한 계약이다.** 전송 수단이 배선되지 않은 엔진에서
critical 알림은 outbox에 쌓이고 게이트는 잠기는데 **루프는 매 주기 아무 말 없이 돈다.**
그 침묵을 깨는 것은 `obs`를 안 고치고도 할 수 있다 — 루프가 `Flush`의 반환값
`(delivered, remaining, err)`에서 *"밀린 것이 있는데 하나도 안 나갔다"*를 읽으면 된다.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 6, returns 4)
- `Flush` 호출 지점 전수: `rg '\.Flush\(' -g '*_test.go' internal/obs/` → 3건
- 프로덕션 호출자 0건: `rg '\.Flush\(' -g '*.go' -g '!*_test.go'` → `bufio`·`http.Flusher`만
