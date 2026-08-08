# Branch Test Map: `ReconcileDriver.alert`

Source: `internal/app/engine/reconcileloop.go` (552-560). AST 기준 분기 2 / 이탈 1 /
defers 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:553` `Alerts == nil` → 무동작 | 간접 — `Alerts` 없이 도는 대사 루프 테스트 다수 | no | yes |
| B2 | `:556` `Notify` 오류 → `alert_undelivered` error 한 줄 | **없음** | no | no |

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R5 | 응답 없는 publisher에서 대사 사이클 1회 | 사이클이 유계 안에 반환한다 |
| R6 | 대사 루프가 critical을 올리는 **동안** exit 루프가 critical을 올린다 | exit 루프가 `n.mu`에서 대기하지 않는다 |

R6이 이 함수를 이 change에 포함시키는 유일한 이유다. 본문은 바꾸지 않지만
**공유 뮤텍스를 통한 간섭**은 계약이 되어야 한다.
