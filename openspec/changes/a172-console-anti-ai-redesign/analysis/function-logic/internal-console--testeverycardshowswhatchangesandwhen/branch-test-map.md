# Branch Test Map: `TestEveryCardShowsWhatChangesAndWhen`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 설정 탭 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B2 | 탭의 카드 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B3 | limits 카드 여부 | 이 함수 | RED: form 속성 의존 | GREEN: matrix 분기 |
| B4 | limits matrix 부재 | 이 함수 | RED: form 속성 의존 | GREEN: matrix 직접 검사 |
| B5 | 일반 카드에 preview 없음 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B6 | 상시 카드 목록 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B7 | effect-timing marker 없음 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B8 | effect-timing class 없음 | 이 함수 | 기존 보호 | GREEN에서 유지 |
