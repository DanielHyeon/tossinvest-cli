# Branch Test Map: TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse (base 이름, 현재 트리에 없음)

- Base source SHA-256: `b765db14f7c1c5b4e1e7fe58b4f808d5128e5dad6355290239f27eb71c7602e8` (`revision: base` — 이 이름의 함수는 현재 트리에 없다).
- 테스트 함수이므로 "이 분기를 덮는 테스트"가 아니라 **이 함수가 사라진 뒤 그 단언을
  누가 이어받는가**를 적는다.

| Branch | Scenario anchor (base) | 이어받은 곳 | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 166:2 — 두 팔 테이블 | 크기 팔은 `TestUnixClientRejectsOversizedResponse` 가 그대로 유지. unknown-field 팔은 `TestClientIgnoresAdditiveFieldsFromANewerEngine` 이 **반대 방향으로** 대체 | 아니오 — 계약이 뒤집힌 것이지 결함이 고쳐진 것이 아니다 | 예 |
| B2 | if at 177:4 — "거절되지 않으면 실패" | 같은 두 테스트. 의미 판정은 `TestClientStillRejectsASemanticallyInvalidSnapshot` 이 유지 | 아니오 | 예 |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M5: `Client.Read` 에 `DisallowUnknownFields()` 를 되돌린다 (제거한 그 동작) | KILLED — `TestClientIgnoresAdditiveFieldsFromANewerEngine` 이 `json: unknown field "coordinators"` 로 실패. 원복 후 `DisallowUnknownFields` 잔여 1건(`readDescriptor`, 의도적으로 유지) 확인. |
