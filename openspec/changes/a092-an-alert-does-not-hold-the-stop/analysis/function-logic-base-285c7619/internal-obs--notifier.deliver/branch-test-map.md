# Branch Test Map: `Notifier.deliver`

Source: `internal/obs/notifier.go` (341-407). AST 기준 **분기 12 / return 3**.

> **12판에서 재기준화했다.** 11판까지의 이 표는 base `ec29dc72`의 좌표(`:238-287`,
> 분기 10)를 썼다. 그 뒤 a096 2라운드와 a097이 같은 파일을 편집해 함수가 **103줄
> 아래로 밀렸고 분기가 둘 늘었다.** 새 base는 `285c7619`이고 아래 좌표는 그것이다.
> 늘어난 둘(B6·B7)은 a096 2라운드가 만든 **publish 성공 + 기록 실패** 경로다.

| Branch | Scenario | Test | 현재 |
|---|---|---|---|
| B1 | `:343` `attempts <= 0` → 기본 3회 | **없음** — 아래 참조 | **미덮임** |
| B2 | `:349` 시도 루프 | `TestPersistentDeliveryFailureBlocksEntries` | 덮임 |
| B3 | `:350` publisher 없음 → 즉시 `break` | **없음** — 아래 참조 | **미덮임** |
| B4 | `:355` publish 성공 | `TestAWiredPublisherDeliversTheCriticalAlert` (`a074_notification_wiring_test.go`) | 덮임 |
| B5 | `:357` `MarkAlertDelivered` 성공 → `return true` `:358` | 같은 테스트 | 덮임 |
| B6 | `:373` 기록 실패 + Log 있음 → 구조화 로그 | `TestASendThatCannotBeRecordedLatchesTheGate` (`a096b_round2_test.go:48`) | 덮임 |
| B7 | `:378` 기록 실패 + Gate 있음 → **`Gate.Block`** → `return false` `:381` | 같은 테스트 | 덮임 |
| B8 | `:384` `MarkAlertAttemptFailed` 오류 → 로그 | **없음** | 미덮임 |
| B9 | `:387` `attempt < attempts` → 대기 | `TestPersistentDeliveryFailureBlocksEntries` (주입 시계) | 덮임 |
| B10 | `:388` ctx 종료 → `break` | **없음** | 미덮임 |
| B11 | `:398` Log 있음 → 소진 기록 | `TestPersistentDeliveryFailureBlocksEntries` — harness가 `Log`를 채운다(`obs_test.go:338`) | 덮임 |
| B12 | `:403` Gate 있음 → **`Gate.Block`** | `TestPersistentDeliveryFailureBlocksEntries` | 덮임 |

## 11판의 무테스트 3개 중 하나는 a096이 닫았다

11판의 표는 B5(구 번호, `MarkAlertDelivered` 오류)를 **"없음 — 그런데 프로덕션에서
6번 발생했다"**로 적었다(2026-08-05 `engine.log`, `journal: no such alert: N (or it is
no longer pending)`). **a096 2라운드가 그 경로를 닫았다** — 지금은 게이트를 잠그고
구조화 로그를 남기며 `TestASendThatCannotBeRecordedLatchesTheGate`가 덮는다.
그것이 지금의 B6·B7이다.

남는 미덮임은 **B1·B3·B8·B10 넷**이다. B8은 시도별 실패 기록의 오류 경로(a089 범위),
B10은 ctx 종료 중 `wait`가 false를 내는 경로다.

> **18라운드 B-P6이 B1과 B3을 `덮임`에서 내렸다.**
>
> **B1**(`attempts <= 0` → 기본 3회)은 `TestPersistentDeliveryFailureBlocksEntries`를
> 이름 붙이고 있었다. 그 테스트는 `Attempts: 3`을 **명시한다**(`obs_test.go:346`) —
> 기본값 대체 분기를 지나지 않는다. `internal/obs`의 `Notifier` 리터럴 중 `Attempts`를
> 채우는 것이 9개이고(`a096:59`·`a096:310`·`a096b:109`·`a097:46`·`a097:143`·
> `escalation:42`·`obs_test:346`·`obs_test:569`·`obs_test:588`), 채우지 않는 셋은
> `Journal`이 없어 `deliver`에 **도달하지 않는다**. `DefaultCriticalAttempts`
> (`notifier.go:45`)는 프로덕션이 실제로 쓰는 값인데 아무 테스트도 통과시키지 않는다 —
> `wait`의 `DefaultRetryDelay`와 **정확히 같은 모양의 구멍**이고, 둘이 함께 34초 예산의
> 두 항을 만든다.
>
> **B3**(`Publisher == nil` → 즉시 `break`)은 `TestCriticalAlertIsDurableBeforeItIsSent`를
> 이름 붙이고 있었다. 그 테스트는 `&failingPublisher{}`를 넣고 발송이 **한 번
> 일어나는 것**을 단언한다(`obs_test.go:353-378`) — publisher가 nil이 아니다.
>
> 두 칸 다 "이 시나리오를 아는 테스트"가 아니라 "이 함수를 지나는 테스트"를 적고
> 있었다. **함수에 도달하는 것과 분기를 타는 것은 다르다.**

## a092가 이 함수에서 쓰는 것

a092는 **이 함수를 편집하지 않는다.** 여기서 읽는 것은 두 가지뿐이다.

- **`attempts`의 출처가 B1이다** — `n.Attempts`가 0이면 `DefaultCriticalAttempts`(3).
  조립부가 채우지 않으므로 오늘의 실효값이 3이라는 근거가 이 분기다.
- **재시도 사이의 대기가 B9→`wait`다** — `wait`의 기본값이 `RetryDelay` 미설정 시
  2초라는 것이 같은 이유로 성립한다.

두 값과 `Ntfy.Timeout`이 a092가 조립부에서 채우는 세 필드다. **`obs` 패키지의 함수는
한 줄도 바꾸지 않는다.**

## 이 change가 요구하는 RED

| # | Scenario | 기대 |
|---|---|---|
| R14 | `deliver`가 호출자 goroutine이 아닌 곳에서 돈다 | B1~B12의 결과가 전부 동일 |
| R15 | 발송 중 두 번째 critical이 들어온다 | `n.mu` 직렬화 유지, 호출자는 대기하지 않는다 |
| R16 | B10 — ctx가 종료된 상태 | `break` 후 게이트 래치. 종료 중 알림이 무한 대기하지 않는다 |

**B8의 무테스트 상태는 이 change가 닫지 않는다**(a089 범위). B10은 R16이 부분적으로만
덮는다. `not-applicable` 사유를 review에 남긴다.
