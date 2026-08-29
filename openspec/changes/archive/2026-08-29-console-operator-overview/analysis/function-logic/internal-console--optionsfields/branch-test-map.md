# Branch Test Map: `optionsFields`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전 파일 순회 | 이 함수의 호출자 둘 | — | yes |
| B2 | 다른 TypeSpec 건너뛰기 | 구조 분기 | — | yes |
| B3 | `Options`가 구조체가 아님 | 타입 교체 변이 | — | n/a |
| B4 | 선언 수가 1이 아님 | positive control — 파일 이름 고정 시절이면 다른 파일 선언에서 0이 나온다 | yes | yes |
