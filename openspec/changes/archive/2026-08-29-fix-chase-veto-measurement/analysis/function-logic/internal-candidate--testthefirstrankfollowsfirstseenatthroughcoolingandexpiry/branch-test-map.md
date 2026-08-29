# Branch Test Map: `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승격 | 자체 실행 | yes (시그니처 변경으로 컴파일 실패) | yes |
| B2 | 첫 기록 | 자체 실행 | yes | yes |
| B3 | 냉각 | 자체 실행 | yes | yes |
| B4 | 냉각은 순위를 보존한다 | 자체 실행 | yes | yes |
| B5 | 재승격 | 자체 실행 | yes | yes |
| B6 | `first_seen_at` 보존 | 자체 실행 | yes | yes |
| B7 | 재진입 기록 시도 | 자체 실행 | yes | yes |
| B8 | 읽기 | 자체 실행 | yes | yes |
| B9 | write-once | 자체 실행 | yes | yes |
| B10 | 상태 읽기 | 자체 실행 | yes | yes |
| B11 | 만료 도달 | 자체 실행 | yes | yes |
| B12 | 만료 후 재승격 | 자체 실행 | yes | yes |
| B13 | 읽기 | 자체 실행 | yes | yes |
| B14 | 만료가 순위를 지운다 | 자체 실행 | yes | yes |
| B15 | instant·source도 지운다 | 자체 실행 | yes | yes |
| B16 | 새 생명의 기록 | 자체 실행 | yes | yes |
| B17 | 새 위치 | 자체 실행 | yes | yes |
| B18 | 새 생명의 백분위 | 자체 실행 | yes (`storedFirstRank`가 자격을 채우지 않으면 미측정이 되어 실패) | yes |

B14·B15는 `Promote`의 reset 절 11개를 확인하지만 이 change가 더한 **두 자격 칼럼**의
초기화는 확인하지 않는다 — 그것은 `TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds`가
결과로 확인한다.
