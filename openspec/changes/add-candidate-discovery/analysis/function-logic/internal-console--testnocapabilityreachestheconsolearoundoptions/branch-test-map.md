# Branch Test Map: `TestNoCapabilityReachesTheConsoleAroundOptions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일 이름 수집 | 이 테스트 자신 | — | yes |
| B2 | 정렬 순회 | 같은 위 | — | yes |
| B3 | 최상위 선언 순회 | 같은 위 | — | yes |
| B4 | 선언 종류 분기 | 같은 위 | — | yes |
| B5 | 함수 선언 | exported `*Console` 메서드 전부 | — | yes |
| B6 | 수신자·exported 판정 | `receiverIsConsole`이 항상 false를 답하는 변이 | yes — B26이 `no exported method on *Console was checked`로 실패 | yes |
| B7 | 메서드 시그니처 closure 동사 검사 | `SetDesk(Desk)` 우회 변이 | yes | yes |
| B8 | 메서드 시그니처의 인터페이스 embed | embed 삽입 변이 | yes | yes |
| B9 | GenDecl | 패키지 var 전부 | — | yes |
| B10 | var가 아닌 선언 건너뛰기 | 구조 분기 | — | yes |
| B11 | spec 순회 | 패키지 var 전부 | — | yes |
| B12 | ValueSpec 아님 | 구조 분기 | — | yes |
| B13 | var 이름 동사 검사 | `var packageDesk Desk` 변이 | yes | yes |
| B14 | 추론 타입 var는 이름만 | **경계 ②** — `var desk = newDesk()` 변이는 이 가드를 통과한다 | no — 문서화된 경계 | n/a |
| B15 | 명시 타입 var의 closure | `var packageDesk Desk` 변이 | yes | yes |
| B16 | 그 closure의 인터페이스 embed | embed 삽입 변이 | yes | yes |
| B17 | 인터페이스 노드만 골라낸다 | `ast.Inspect` 콜백이 첫 노드에서 `false`를 답하는 변이 | yes — B22·B25 실패 | yes |
| B18 | 방문 기록 뒤 seam 제외 | 같은 변이(방문 기록이 비면 B25가 `6개 중 6개`로 실패) | yes | yes |
| B19 | 그 메서드 목록 순회 | 인라인 `interface{ PlaceOrder(…) }` 변이 | yes | yes |
| B20 | 그 메서드 이름 동사 검사 | 같은 위 | yes | yes |
| B21 | seam을 하나도 못 찾음 | positive control — `Options` 걷기 | yes | yes |
| B22 | 인터페이스 선언에 아예 닿지 못함 | positive control — `ast.Inspect` 콜백 즉시 `false` 변이 | yes — `the package-wide walk reached no interface declaration at all` | yes |
| B23 | 미방문 seam 계수 순회 | 아래 B24·B25와 같은 변이 | yes | yes |
| B24 | seam 하나가 방문 목록에 없음 | 파일 하나(`holdings.go`)를 걷기에서 조용히 빼는 변이 | yes | yes |
| B25 | 미방문 seam이 하나라도 있음 | 같은 변이 | yes — `1 of the 6 Options seams were never visited` | yes |
| B26 | `*Console` 메서드를 하나도 검사하지 못함 | positive control — `receiverIsConsole` 항상 false 변이 | yes — `the method walk is vacuous` | yes |
