# Branch Test Map: `Notifier.Notify`

Source: `internal/obs/notifier.go` (124-133). AST 기준 분기 1 / 이탈 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:128` normal → `publishBestEffort`, `return nil` `:130` | `TestOrdinaryAlertsAreBestEffort` (`obs_test.go:527`) | no | yes |
| — | critical → `notifyCritical`, `return` `:132` | `TestCriticalAlertIsDurableBeforeItIsSent` (`obs_test.go:353`) · `TestPersistentDeliveryFailureBlocksEntries` (`:384`) | no | yes |

> **18라운드 B-P2가 지운 것.** 이 표의 B1은 `TestMeasurementEventsAreNeverCritical`
> (`measurement_test.go:31`)도 함께 이름 붙이고 있었다. 그 테스트는 `Notifier`를
> 만들지도 `Notify`를 부르지도 않는다 — `Subject()`·`SeverityOf()`·`CriticalEvents()`만
> 부르는 등급표 테스트다(`measurement_test.go:31-55`). 남은 `TestOrdinaryAlertsAreBestEffort`
> 하나는 `EventFillObserved`로 실제 `n.Notify`를 타므로 유효하다.

## 무엇이 단언되지 않는가

두 이탈 다 **결과**는 단언되고 **소요 시간**은 단언되지 않는다.
`TestPersistentDeliveryFailureBlocksEntries`는 재시도 소진 후 게이트 래치를 확인하지만
주입 시계(`Clock`)를 쓰므로 실제 벽시계 체류를 재지 않는다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R7 | critical 이벤트 1건, publish가 블록 | `Notify`가 outbox 쓰기 직후 반환한다 |
| R8 | R7 뒤 발송이 결국 성공 | outbox 행이 DELIVERED가 된다 — 비동기화가 durability를 깨지 않는다 |
| R9 | R7 뒤 발송이 예산까지 실패 | 게이트가 래치되고 ENTRY_BLOCKED로 강화된다 — 현행 결과가 보존된다 |
| R10 | normal 이벤트, publish가 블록 | `Notify`가 즉시 반환하고 발송은 나중에 시도된다(최선노력 계약 유지) |
