# Branch Test Map: `Console.writeQuarantineError`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 오류 종류로 갈린다 | 기존 a079 표 테스트 (`a079_quarantine_release_test.go`) | no | yes |
| B2 | ErrUnwired → 501 + 강등 가능성을 말하는 문구 | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` · `TestTheUnwiredQuarantineMessageDoesNotBlameTheBuild` | no — a109 편집과 같은 커밋에서 도입한 핀 | yes |
| B3 | ErrNotQuarantined → 404 | 기존 a079 표 테스트 | no | yes |
| B4 | ErrVersionMismatch → 412 | 기존 a079 표 테스트 | no | yes |
| B5 | ErrCapabilityTooEarly → 425 | 기존 a079 표 테스트 | no | yes |
| B6 | ErrCapabilityExpired → 410 | 기존 a079 표 테스트 | no | yes |
| B7 | ErrCapabilityInvalid → 403 | 기존 a079 표 테스트 | no | yes |
| B8 | ErrConfirmationRequired → 400 | 기존 a079 표 테스트 | no | yes |
| B9 | 그 밖 → 400 + 원문 | 기존 a079 표 테스트 | no | yes |
