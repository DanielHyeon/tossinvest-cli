# Branch Test Map: `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `/positions`의 재로드 주기가 기대와 다르다 | 갱신 후 동일 함수 | yes (변이 6.1) | yes |
| B2 | `/history`가 자동 재로드를 갖게 되었다 | 갱신 후 동일 함수 | no | yes |
| — | 열린 탭의 브로커 비용 상한 (base에서 B1이 대리하던 것) | `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling` | yes (변이 6.2 — 5 허용에 24 관측) | yes |

세 번째 행이 이 편집의 요점이다. base는 상한을 주기로 대리 판정했고 그 대리는
부정확했다. a080은 호출 수를 직접 세는 테스트로 옮겼고, 변이 6.2가 그 테스트의
실효성을 확인했다.
