# Branch Test Map: `TestTheDashboardScreensAreReads`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표 전체 순회 | 이 테스트 자신 | — | yes |
| B2 | 관심 밖 경로 건너뛰기 | 같은 위 | — | yes |
| B3 | 읽기 화면이 CSRF 게이트 뒤에 있으면 실패 | `mutating` 추가 변이 | yes | yes |
| B4 | 세션 게이트 누락 | `session0` 제거 변이 | yes | yes |
| B5 | 다섯 화면 존재 확인 순회 | 이 테스트 자신 | — | yes |
| B6 | 화면이 등록되지 않았으면 실패 | 등록 제거 변이 + 추출기 파일 범위 축소 변이(세 화면이 한꺼번에 사라진다) | yes | yes |
