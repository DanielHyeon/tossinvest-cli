# Branch Test Map: `checkCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 도달 가능한 모든 타입 이름의 동사 검사 | `AccountHandle`→`OrderHandle` 개명 변이가 통과했다는 것이 이전 가드의 반증이었고, 지금은 **열어 본다** | yes | yes |
| B2 | 도달 가능한 모든 인터페이스의 embed 검사 | embed 삽입 변이 | yes | yes |
| B3 | 인터페이스가 아닌 필드 | func 타입 seam 8 + 평문 10 | — | yes |
| B4 | 메서드가 열거됐는데 인터페이스로 해석되지 않음 | 별칭 사슬 변이 | yes | yes |
| B5 | 메서드 집합 없음을 적극적으로 읽을 수 없는 모양 | `Feed any` 변이 | yes | yes |
| B6 | 빈 인터페이스 seam | 빈 인터페이스 변이 | yes | yes |
| B7 | seam 메서드 수집 | 인터페이스 seam 6 | — | yes |
| B8 | embed는 여기서 보고하지 않는다 | `checkNoEmbedding`이 맡는다 | — | yes |
| B9 | 메서드 이름 수집 | 인터페이스 seam 6 | — | yes |
| B10 | 허용 집합 구성 | 같은 위 | — | yes |
| B11 | 선언 순회 | 같은 위 | — | yes |
| B12 | 허용 밖 메서드 | 두 번째 메서드 추가 변이(`TestTheSettingsSeamDeclaresExactlyLoadAndSave`가 같은 형태를 별도로 고정) | yes | yes |
| B13 | 허용 집합 순회 | 같은 위 | — | yes |
| B14 | 약속한 메서드 상실 | 메서드 제거 변이 | yes | yes |
