# Branch Test Map: `Console.mutating`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승인 라우트에 GET | `TestTheApprovalRoutesRefuseAGET` | — (동작 무변경) | yes |
| B2 | 폼을 읽을 수 없는 요청 | 직접 테스트 없음 — 무변경 분기이며 fail-closed 방향이다 | — (동작 무변경) | n/a |
| B3 | 세션은 유효하나 CSRF가 틀린 POST | `TestAWrongCSRFTokenSendsNothing`, `TestTheEngineButtonsNeedTheSessionAndTheCSRFToken` | — (동작 무변경) | yes |
