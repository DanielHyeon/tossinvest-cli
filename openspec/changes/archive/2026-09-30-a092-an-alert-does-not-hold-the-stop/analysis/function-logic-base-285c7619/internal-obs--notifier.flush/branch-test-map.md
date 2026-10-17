# Branch Test Map: `Notifier.Flush`

Source: `internal/obs/notifier.go` (427-462). AST 기준 분기 6 / 이탈 4 /
defers 1 / go_statements 0.

**a092가 편집하는 함수**이므로 RED 열이 실제 의무를 담는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:428` journal이 nil → 무동작 | **없음** | no | no |
| B2 | `:438` `PendingAlerts` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| B3 | `:441` 밀린 행을 순회해 전송한다 · **그리고 상한 없이 전부 순회한다**(17판이 고칠 것) | 순회: `internal/obs/obs_test.go TestRecoveredDeliveryDoesNotReleaseTheGateByItself:427`(`:440`에서 호출), `TestCriticalAlertSurvivesAProcessRestart:551`(`:590`) · **상한: 없음 → R17-5** | 순회 no / 상한 **예정** | 순회 yes / 상한 **예정** |
| B4 | `:442` **`Publisher == nil` → `break`, 시도 기록 없음** | **없음** → **R17-4** | **예정** | **예정** |
| B5 | `:451` `Publish` 실패 → 시도 기록 후 다음 행 | **없음.** `Flush`를 부르는 테스트는 셋뿐이고(`obs_test.go:440`·`:590`, `a097_exclusion_is_an_event_test.go:109`) **셋 다 전송이 성공하는 구성**이다 — `:440`은 직전에 `pub.setFail(false)`, `:590`은 `recovered := &failingPublisher{}`(기본 성공), `:109`는 released reentrant publisher | no | no |
| B6 | `:455` `MarkAlertDelivered` 실패 → 루프 중단 | **없음** — 이미 `DELIVERED`인 행을 `Flush`가 만나는 테스트가 없다 | no | no |
| — `:461` | 남은 수를 세어 돌려준다 | `obs_test.go:440`·`:590`이 `remaining`을 검사한다 | no | yes |
| — 잠금 | `n.mu`를 잡아 발송과 배제된다 | `a097_exclusion_is_an_event_test.go TestFlushCannotPublishBesideASend:92` | yes (a097) | yes |

B3는 같은 분기의 **두 성질**(순회한다 / 상한이 없다)을 한 행에 담았다.
`check_analysis.py`가 분기 ID 중복을 거부하므로 한 분기는 한 행이다.
마지막 행은 AST 분기가 아니라 `:434-435`의 잠금에 대한 관측이므로
`B` ID를 붙이지 않는다.

## a092가 이 함수에 대해 지는 RED

**"예정"은 아직 관측하지 않았다는 뜻이다.** 이 표에 `yes`를 적는 것은
그 테스트를 실제로 돌려 실패를 본 뒤다 — 지금 적으면 그것이
"측정하지 않은 커버리지 주장"이 된다.

| RED | 무엇을 관측하나 | 이 표의 어느 행 |
|---|---|---|
| R17-1 | 관측 사이클이 전송 응답을 기다리지 않는다 | B3 전체 |
| R17-2 | 기록 경로가 배달 잠금을 기다리지 않는다 | 마지막 행의 **반대**를 요구한다 |
| R17-4 | nil publisher도 실패한 시도로 세어진다 | B4 |
| R17-5 | 한 주기가 처리하는 행 수에 상한이 있다 | B3 두 번째 행 |
| R17-6 | 한 주기가 한 행에 시도를 한 번만 쓴다 | B5·B6 |

**R17-2가 `TestFlushCannotPublishBesideASend:92`와 정면으로 부딪힌다.**
그 테스트는 *"`Flush`가 발송 옆에서 발행할 수 없다"*를 `n.mu`로 보장하고,
17판은 그 배제를 뮤텍스가 아니라 SQL CAS로 옮긴다. **기존 테스트를 지우지 않고
성질을 다시 표현해야 한다** — 지우면 a096 라운드 1의 이중 발송이 되돌아온다.
§8 GREEN의 조건으로 tasks.md에 남긴다.

미테스트 B1·B2·B6은 a092가 편집하는 함수의 분기이므로 **면제 대상이 아니다.**
B6은 특히 17판에서 도달 가능성이 오른다(배달 루프가 주기마다 돌므로).
§6.0에 **R17-13**으로 세운다 — 이 산출물이 만든 항목이다.
