# Branch Test Map: `ExitObserver.clearDelay`

AST 기준 분기 0. 조건 없는 삭제이므로 분기 테스트 대상이 아니다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — 무조건 두 map에서 키 삭제 | `TestAnUncancellableEntry...` `:914-919` (취소 성공 후 청산이 나간다) | no | yes |

## 필요한 RED — 호출 지점에 대해

| # | Scenario | 기대 |
|---|---|---|
| R1 | 살아 있는 주문 없이 보호 제출이 거부됨 | `record:1150`이 시계를 지우지 **않는다** (1라운드 C1 회귀) |
| R2 | working order 정리 실패 후 취소가 성공 | 시계가 해제된다 — `:914-919` 무변화 |
