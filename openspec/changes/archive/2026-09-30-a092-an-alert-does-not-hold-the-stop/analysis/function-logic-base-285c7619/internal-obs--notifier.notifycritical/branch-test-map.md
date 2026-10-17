# Branch Test Map: `Notifier.notifyCritical`

Source: `internal/obs/notifier.go` (170-225). AST 기준 분기 4 / 이탈 3 / defers 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:171` journal 없음 → 최선노력 강등, `return nil` `:181` | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` (`obs_test.go:600`) | no | yes |
| B2 | `:175` journal 없음 + Log 있음 → 경고 한 줄 | 같은 테스트 | no | yes |
| B3 | `:195` `claimAndDeliver` 오류 → `escalate` `:213` + `return` 오류 `:214` | `TestAClaimThatFailsBlocksNewEntries` (`a097_claim_failure_blocks_entry_test.go:61`) · `TestAClaimThatFailsAttemptsTheDurableBlock` (`:93`) · `TestAFailedClaimStillReturnsItsError` (`:113`) · `TestAFailedClaimWithNothingWiredStillReports` (`:136`) | no | **yes** |
| B4 | `:217` `owed && !sent` → `escalate` `:222` | `TestPersistentDeliveryFailureBlocksEntries` (`obs_test.go:384`) · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` (`measurement_test.go:145`) | no | yes |
| — | 정상 `return nil` `:224` | `TestCriticalAlertIsDurableBeforeItIsSent` (`obs_test.go:353`) | no | yes |

> **18라운드가 이 표를 양방향으로 고쳤다.** 좌표는 a097 이후 재고정된 적이 없었고
> (`(153-190)` → 실제 `(170-225)`), **B3의 커버리지는 과소 보고되어 있었다.**
> 17판까지 B3은 `없음`·GREEN `no`였다. a097이 저널을 닫아 claim을 실패시키는 테스트
> **넷**을 추가했고 그 넷이 전부 `Notify`가 오류를 반환하는 것을 단언한다
> (`a097_claim_failure_blocks_entry_test.go:65-66`·`:97-98`·`:117-118`·`:150-152`).
>
> 거짓 GREEN만 위험한 것이 아니다. **거짓 `no`는 이미 있는 테스트를 다시 쓰게 만든다** —
> 아래 R11이 그렇게 태어났다.

## 순서를 단언하는 테스트는 하나뿐이다

`TestCriticalAlertIsDurableBeforeItIsSent`(`obs_test.go:353`)의 이름 자체가 계약이다 —
**기록이 발송보다 먼저**. 이 change는 그 순서를 바꾸지 않고 **발송만 뒤로 민다.**
그래서 이 테스트는 GREEN을 유지해야 하고, 유지하지 못하면 설계가 틀린 것이다.

`TestCriticalAlertSurvivesAProcessRestart`(`obs_test.go:551`)는 재기동 후 `Flush`가
배달하는 것을 단언한다 — **비동기 발송이 이미 이 저장소에 있는 개념**이라는 증거다.
없는 것은 그 `Flush`를 부르는 프로덕션 호출자다.

## 필요한 RED

| # | Scenario | 기대 | 상태 |
|---|---|---|---|
| R11 | B3 — claim이 오류 | `Notify`가 오류를 반환한다(기록 실패만이 호출자 오류) | **born green — 철회.** a097의 네 테스트가 이미 단언한다 |
| R12 | 발송이 예산까지 실패 → escalate·게이트 래치가 **여전히** 일어난다 | B4의 두 부작용이 비동기 경로에서도 보존 | 유효 |
| R13 | 발송 goroutine이 다시 critical을 올린다 | 교착하지 않는다(주석 `:210-212`·`:218-221`의 재진입 위험) | 유효 |

**R11을 철회하는 이유는 통과해서가 아니라 이미 있어서다.** 새 RED가 기존 테스트와
같은 것을 단언하면 그것은 증거를 늘리지 않고 유지 비용만 늘린다. 이 change가 B3에 대해
새로 져야 할 것은 **배달 실행자가 죽었을 때도 같은 결과가 나오는가**이고, 그것은
R18-1이 진다.
