# Branch Test Map: `Notifier.notifyCritical`

Source: `internal/obs/notifier.go` (153-190). AST 기준 분기 4 / 이탈 3 / defers 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:154` journal 없음 → 최선노력 강등, `return nil` `:164` | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` (`obs_test.go:600`) | no | yes |
| B2 | `:158` journal 없음 + Log 있음 → 경고 한 줄 | 같은 테스트 | no | yes |
| B3 | `:178` `EnqueueAlert` 오류 → `return` 오류 `:179` | **없음** | no | no |
| B4 | `:182` `deliver` 실패 → `escalate` `:187` | `TestPersistentDeliveryFailureBlocksEntries` (`:384`) · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` (`measurement_test.go:145`) | no | yes |
| — | 정상 `return nil` `:189` | `TestCriticalAlertIsDurableBeforeItIsSent` (`:353`) | no | yes |

## 순서를 단언하는 테스트는 하나뿐이다

`TestCriticalAlertIsDurableBeforeItIsSent`(`:353`)의 이름 자체가 계약이다 —
**기록이 발송보다 먼저**. 이 change는 그 순서를 바꾸지 않고 **발송만 뒤로 민다.**
그래서 이 테스트는 GREEN을 유지해야 하고, 유지하지 못하면 설계가 틀린 것이다.

`TestCriticalAlertSurvivesAProcessRestart`(`:551`)는 재기동 후 `Flush`가 배달하는 것을
단언한다 — **비동기 발송이 이미 이 저장소에 있는 개념**이라는 증거다. 없는 것은
그 `Flush`를 부르는 프로덕션 호출자다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R11 | B3 — `EnqueueAlert`가 오류 | `Notify`가 오류를 반환한다(기록 실패만이 호출자 오류) |
| R12 | 발송이 예산까지 실패 → escalate·게이트 래치가 **여전히** 일어난다 | B4의 두 부작용이 비동기 경로에서도 보존 |
| R13 | 발송 goroutine이 다시 critical을 올린다 | 교착하지 않는다(주석 `:181-186`의 재진입 위험) |
