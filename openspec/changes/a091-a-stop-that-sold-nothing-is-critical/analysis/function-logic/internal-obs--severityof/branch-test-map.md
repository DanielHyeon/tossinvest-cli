# Branch Test Map: `SeverityOf`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:310` `criticalEvents`에 있으면 critical | `TestAQuarantineCreationIsCritical` `internal/obs/a074_quarantine_event_test.go:16` · `exitloop_test.go:910`,`:1013` | no | yes |

기본 경로(`:313`, normal)는 분기 id가 없다. `TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`
(`exitloop_test.go:478`)가 normal 등급을 확인한다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 새로 승격되는 이벤트 종류 | `SeverityOf`가 critical을 돌려준다 |
| R2 | 기존 18종 | **무변화** — 등급 회귀 0 |
| R3 | 미등록 종류 | 여전히 normal (기본값 보존) |
