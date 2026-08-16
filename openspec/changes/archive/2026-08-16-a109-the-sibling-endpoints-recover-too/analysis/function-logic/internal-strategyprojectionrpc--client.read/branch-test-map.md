# Branch Test Map: `Client.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 반쪽 client(토큰 짧음·transport 없음) | 커버 없음 — `Dial`만 이 타입을 만들고 셋을 함께 채운다 | no | no |
| B2 | 요청 생성 실패 | 커버 없음(경로가 상수라 만들 수 없다) | no | no |
| B3 | endpoint가 죽었다 | a109 `TestTheDaemonReattachesAfterTheEngineRestarts` | yes(§2.3) | yes |
| B4 | 본문 읽기 실패 | 커버 없음(정상 endpoint에서 만들 수 없다) | no | no |
| B5 | 응답이 상한을 넘는다 | `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse` | no(a108 소유) | yes(회귀) |
| B6 | 200이 아닌 응답 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no(a108 소유) | yes(회귀) |
| B7 | 알 수 없는 필드·깨진 JSON | `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse` | no(a108 소유) | yes(회귀) |
| B8 | JSON 값이 둘 이상 | 같은 테스트 | no(a108 소유) | yes(회귀) |
| 종단 | 정상 스냅샷 | `TestAuthenticatedUnixProjectionCrossNamespaceAndReadOnlyAuthority` · a109 `TestTheClientCanBeLetGo`(Close 뒤 재읽기) | no | yes |
