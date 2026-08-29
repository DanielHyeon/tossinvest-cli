# Branch Test Map: `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 라운드 2회 — 새로고침해도 해석이 늘지 않는다 | 이 테스트 | yes | yes |
| B2 | 읽기 화면 2종 순회 | 동일 | no (구조) | yes |
| B3 | 화면 열기 실패 보고 | 동일 | no (진단) | yes |
| B4 | **모든 읽기 화면을 열어도 계좌 해석 1회** | 동일 | yes — seam별 resolver에서 2회 관측 | yes (1회) |
| B5 | 동시 개시 goroutine 기동 | 동일 | no (구조) | yes |
| B6 | 화면별 호출 분기 | 동일 | no (구조) | yes |
| B7 | 포지션 화면 동시 개시 | 동일 | no (구조) | yes |
| B8 | /orders 동시 개시 | 동일 | no (구조) | yes |
| B9 | 동시 개시 에러 수집 | 동일 | no (진단) | yes |
| B10 | 동시 개시 에러 보고 | 동일 | no (진단) | yes |
| B11 | **두 화면 동시 개시도 해석 1회 — 락이 직렬화** | 동일(`-race`) | yes (seam별 resolver에서 2회) | yes |
| B12 | `console.go` 파싱 실패 시 가드가 스스로 실패 | 동일 | no (자기 진단) | yes |
| B13 | 선언 순회 | 동일 | no (구조) | yes |
| B14 | 함수 선언만 검사 | 동일 | no (구조) | yes |
| B15 | 수신자 있는 메서드 이름 조립(`consoleBroker.resolve`) | 동일 | yes (수신자를 빼면 논증 map과 불일치로 FAIL) | yes |
| B16 | 호출 노드만 검사 | 동일 | no (구조) | yes |
| B17 | `verifyBrokerFactory` 호출 집계 | 동일 | yes | yes |
| B18 | 수집된 구축 지점 순회 | 동일 | no (구조) | yes |
| B19 | 논증되지 않은 새 구축 지점은 FAIL | 동일 | yes (seam에 factory 호출을 심으면 FAIL) | yes |
| B20 | 논증 map 순회 | 동일 | no (구조) | yes |
| B21 | 더 이상 구축하지 않는 예외 항목도 FAIL | 동일 | yes (map에 가짜 항목 추가 시 FAIL) | yes |
| B22 | 선언 순회(`runConsole` 탐색) | 동일 | no (구조) | yes |
| B23 | `runConsole`만 검사 | 동일 | no (구조) | yes |
| B24 | 대입문/호출식 분기 | 동일 | no (구조) | yes |
| B25 | 대입문 검사 | 동일 | no (구조) | yes |
| B26 | 대입문 우변 순회 | 동일 | no (구조) | yes |
| B27 | 호출이 아닌 우변은 건너뛴다 | 동일 | no (구조) | yes |
| B28 | `newConsoleBroker` 호출만 센다 | 동일 | yes | yes |
| B29 | 좌변 식별자를 공유 resolver 이름으로 기록 | 동일 | yes (기록 없이는 B37이 비교할 값이 없다) | yes |
| B30 | 호출식 검사 | 동일 | no (구조) | yes |
| B31 | 인자 1개짜리 호출만 검사 | 동일 | no (구조) | yes |
| B32 | seam 생성자 이름 순회 | 동일 | no (구조) | yes |
| B33 | seam 인자 표현식 기록 | 동일 | yes | yes |
| B34 | **`runConsole`이 공유 resolver를 정확히 1개 만든다** | 동일 | yes (두 번째 생성 시 FAIL) | yes |
| B35 | seam 이름 순회 | 동일 | no (구조) | yes |
| B36 | seam 배선이 사라지면 FAIL | 동일 | yes (배선 삭제 시 FAIL) | yes |
| B37 | **모든 읽기 seam이 그 식별자를 받는다** | 동일 | yes (`consoleOrdersSeam(newConsoleBroker(root))`로 바꾸면 FAIL) | yes |
