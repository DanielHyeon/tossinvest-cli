# Branch Test Map: `Console.routes`

a063은 분기를 추가하지 않았다. 재생성한 AST도 분기 2개 그대로이고(B2만 라우트 두
줄만큼 아래로 밀렸다), 늘어난 것은 `mux.HandleFunc` 호출 두 개다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 신뢰되지 않은 원격에서만 로그인 라우트를 등록한다 | 기존 `remote_test.go` 계약 테스트 | no | pass |
| B2 | 원격이면 security 래핑을 씌워 반환한다 | 기존 `remote_csp_test.go` | no | pass |

분기가 아닌 계약(같은 함수에서 지켜야 하는 것): 격리 preview/apply 두 라우트가
`session0(mutating(...))` 아래에 있어야 한다. 이는
`TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 검사하며, 두 경로를
state-changing 목록에 등재하지 않은 상태에서 **실제로 실패했다**(2026-08-04) —
즉 이 계약은 문서가 아니라 게이트다.
