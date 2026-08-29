# Branch Test Map: `methodless`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 타입 종류 분기 | 전 필드 | — | yes |
| B2 | func 타입 | func 타입 seam 8 | — | yes |
| B3 | 구조체는 데이터 | 구조체 필드 | — | yes |
| B4 | 식별자 | 평문 10 | — | yes |
| B5 | builtin 철자 | `Port int` 등 | — | yes |
| B6 | 순환 | 순환 별칭 변이 | — | yes |
| B7 | 선언으로 재귀 | 별칭 사슬 변이 | yes | yes |
| B8 | 다른 패키지 타입은 근거가 적힌 것만 | `Out io.Writer` + 미등록 한정 타입 변이 | yes | yes |
| B9 | 포인터 | 현재 `Options`에 없음(방어) | — | n/a |
| B10 | 괄호 | 현재 표면에 없음(방어) | — | n/a |
| B11 | 슬라이스 | `RequiredEndpoints []string` | — | yes |
| B12 | 맵 | 현재 표면에 없음(방어) | — | n/a |
| B13 | 채널 | 현재 표면에 없음(방어) | — | n/a |
