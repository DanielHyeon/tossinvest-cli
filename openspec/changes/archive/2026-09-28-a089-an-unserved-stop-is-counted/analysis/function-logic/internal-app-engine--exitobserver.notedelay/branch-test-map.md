# Branch Test Map: `ExitObserver.noteDelay`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1572` 시계 미가동 → 시작만 하고 반환 | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` `:883` 1차 관측 | no | yes |
| B2 | `:1576` 한계 내이거나 이미 발화 → 무동작 | 같은 테스트 `:900` (한계 내 알림 0건) | no | yes |

한계 초과 경로(`:1579`)는 분기 id가 없지만 같은 테스트 `:904-912`가 확인한다 —
31초 경과 후 `EventExitLiquidationDelayed`가 critical로 1건.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 같은 원인으로 한계를 여러 배 초과 | 알림은 **1회** (`delayAlerted` latch) |
| R2 | 해제 후 재발 | 시계가 다시 시작되고 다시 발화 가능 |
