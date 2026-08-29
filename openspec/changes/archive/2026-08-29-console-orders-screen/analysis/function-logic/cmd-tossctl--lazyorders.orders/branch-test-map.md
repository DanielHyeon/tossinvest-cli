# Branch Test Map: `lazyOrders.Orders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `resolve()` 실패 — 읽을 것이 아예 없다 | `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`의 factory 대체 + `Orders` 에러 반환 경로 | yes | yes |
| B2 | 브로커가 raw 읽기를 갖지 않는 빌드 | `%T` 메시지 경로 (컴파일 타임 형 계약 + 형상 테스트로 고정) | no (현재 빌드에서 도달 불가 — 형 계약이 앞선다) | yes (형 계약 유지 확인) |
| B3 | OPEN 콜 실패 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately` (대칭 케이스) | yes | yes |
| B4 | OPEN 콜 성공 — `hasNext=false`이므로 건수는 하한이 아니라 수 | `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole` | yes (status 미전송으로 400) | yes |
| B5 | CLOSED 콜 실패 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately` | yes | yes |
| B6 | CLOSED 콜 성공 — `hasNext=true`가 `ClosedTruncated`로 살아남는다 | 동일 + `TestOneRefreshAsks...` | yes (hasNext 유실 시 확정 건수로 렌더) | yes |
| B7 | 조건주문 429 — 나머지 둘은 살아남고 `ConditionalError`가 채워진다 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately` | yes (에러를 삼키면 조건주문 0건이 측정값처럼 렌더) | yes |
| B8 | 조건주문 성공 | `TestOneRefreshAsks...` (빈 목록 200) | yes | yes |
| B9 | 조건주문 레코드 변환 — 원문 보존 | 동일 | yes | yes |

세 번 새로고침해도 클라이언트 구축은 1회(같은 테스트), 포지션 화면까지 열어도 계좌 해석은 1회(`TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`)다.
