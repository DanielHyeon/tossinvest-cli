# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

분기 여섯 그대로다. a065는 조건을 바꾸지 않고 `stateChanging` map에 셋을 더했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록된 모든 라우트를 순회한다 | 자기 자신 | no | pass |
| B2 | 각 라우트를 분류한다 | 자기 자신 | no | pass |
| B3 | 목록에 있는데 CSRF 게이트가 없다 | 자기 자신 | no | pass |
| B4 | CSRF 게이트가 있는데 목록에 없다 | 자기 자신 | **yes** | pass |
| B5 | 목록 전체를 순회한다 | 자기 자신 | no | pass |
| B6 | 목록에 있는데 등록되지 않았다 | 자기 자신 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 스펙 동기화 | 상태변경 목록과 스펙 문장이 같은 커밋에서 움직였다 | `openspec validate --all --strict` + 이 검사 | no | pass |
