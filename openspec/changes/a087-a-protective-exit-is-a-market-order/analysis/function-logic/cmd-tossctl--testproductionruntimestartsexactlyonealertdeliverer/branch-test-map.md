# Branch Test Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 조립 오류 | 기존 runtime 시험 | 기존 green | 오류 반환 유지 |
| B2 | `exit-lower-limit-refresh` 누락·중복 | 이 함수 | 갱신기 추가 전 기대집합 | 전체 집합 동일성 |
| success | 휴면 보조 실행자 포함 | 이 함수 | 해당 없음 | 보조 실행자 다섯 개 |
