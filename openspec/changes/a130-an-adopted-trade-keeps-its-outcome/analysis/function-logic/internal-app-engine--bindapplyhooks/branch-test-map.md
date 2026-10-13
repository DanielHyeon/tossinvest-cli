# Branch Test Map: `bindApplyHooks`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:188` SetApplyHooks 오류 | TestTheProjectionGuardAsksTheBindingAndNotTheData(두 번째 바인딩 거절, interlock_internal_test.go:202) | no | no |

## a130 시나리오 (2판)

- 성공 — 인자 없으면 싱크 없음 · 싱크 인자면 있음 — `TestA130TheLoggedBindingCarriesTheSink` · TestBindingWiresBothApplyHooks
- B1 — 두 번째 바인딩 거절 — 기존 interlock_internal_test
