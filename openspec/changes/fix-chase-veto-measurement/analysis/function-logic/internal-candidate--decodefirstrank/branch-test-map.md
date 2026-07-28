# Branch Test Map: `decodeFirstRank`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기록되지 않은 위치 | `TestARankFromOutsideTheIdentityWindowIsNotStored` · `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows`(late 후보) | — (기존 동작) | yes |
| B2 | 요청 행 수가 있는 행과 없는 행 | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext`(없음) · `TestATruncatedReadingReachesTheVerdictAsTruncated`(있음) | yes | yes |
| B3 | instant 있는 행 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | — (기존 동작) | yes |
| B4 | 읽을 수 없는 stamp | **커버 없음** — 손상된 행 | no | no |

`newly`의 **미인식 문자열** arm(NULL도 'yes'도 'no'도 아닌 값)은 칼럼 CHECK가 막으므로
정상 경로로는 만들 수 없고, 전용 테스트도 없다. `newlyListedFromStore`의 단위 테스트가
`internal/candidate/newlylisted_test.go`에 있다.
