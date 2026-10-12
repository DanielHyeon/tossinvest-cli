# Branch Test Map: `bindApplyHooks`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:188` SetApplyHooks 오류 | TestTheProjectionGuardAsksTheBindingAndNotTheData(두 번째 바인딩 거절, interlock_internal_test.go:202) | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1 — 두 번째 바인딩 거절 — 기존 interlock_internal_test(계속 green — 위임 뒤에도 같은 오류)
- fallthrough — 성공 바인딩 — TestBindingWiresBothApplyHooks + a130 바인딩 시험(싱크 유무)
