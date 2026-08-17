# Branch Test Map: `Console.handleQuarantineReleaseApply`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 미배선이면 501 + 강등 가능성을 말하는 문구 | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` | no — a109 편집과 같은 커밋에서 도입한 핀 | yes |
| B2 | capability 가 비면 403 | 기존 a079 테스트 (`a079_quarantine_release_test.go`) | no | yes |
| B3 | 해제 실패는 writeQuarantineError 로 | 기존 a079 테스트 | no | yes |
