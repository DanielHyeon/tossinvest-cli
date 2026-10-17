# Branch Test Map: `ExitObserver.alertProposalRefused`

**AST branches 0 · returns 0.** 매핑할 분기가 없다는 것이 이 표가 고정하는 사실이다.
`check_analysis.py:300`의 규약대로 분기 없는 함수는 happy-path 한 행(`B1`)으로 적는다.
a092는 이 함수를 편집하지 않는다.

| Branch | 위치 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1550` | happy path — 조건 없이 곧장 `o.alert`. 억제도 조기 반환도 없으므로 **호출 1회가 `Notify` 1회**다 | 기존 커버리지(편집 없음) | n/a | n/a |

## 이 표가 반증하는 주장

| 주장 | 판정 | 근거 |
|---|---|---|
| "ObserveOnce 1회의 동기 체류 ≤ 예산" | **거짓** | 이 함수에 억제가 없어 사이클당 `Notify`가 여러 번 일어난다 |
| "알림 하나의 동기 체류 ≤ 예산" | 참 | a092가 정하는 것이 이것이다 |
