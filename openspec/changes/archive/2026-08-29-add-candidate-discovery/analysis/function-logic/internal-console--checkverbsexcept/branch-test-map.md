# Branch Test Map: `checkVerbsExcept`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 근거가 적힌 다섯 철자만 통과 | `TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated`(여섯 번째 이름·두 번째 seam·근거 없는 예외 셋을 모두 실패시킨다) | yes | yes |
| B2 | 동사 목록 순회 | 전 이름 | — | yes |
| B3 | 동사 철자 포함 | `PlaceOrder` 추가 변이 + 같은 테스트의 `order`·`conditional` 목록 잔존 확인 | yes | yes |
