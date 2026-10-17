# Branch Test Map: `ExitObserver.alertRefused`

a092는 이 함수를 편집하지 않는다. 표는 **인용한 분기가 실재함**을 AST로 고정한다.

| Branch | 위치 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1517` | 같은 포지션의 두 번째 판정 거부는 알리지 않는다 — **래치** | 기존 커버리지 | n/a | n/a |
| B2 | `:1523` | `RefusalError`면 `invariant` 필드를 채운다 | 기존 커버리지 | n/a | n/a |

## 짝 비교 — 같은 판정 경로, 다른 억제

| 함수 | AST branches | 래치 | 사이클당 `Notify` |
|---|---|---|---|
| `alertRefused` `:1516` | 2 | **있음** `o.refused[ID]` | 포지션당 1회(래치 유지 동안) |
| `alertProposalRefused` `:1548` | **0** | **없음** | **포지션마다 매번** |
