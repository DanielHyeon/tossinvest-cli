# Branch Test Map: `BuildProgress`

- Source: `internal/verifylive/report.go` (315-343)
- AST branches 7
- 측정: base `de147cc2` 에서 시험별 `go test -run '^T$' -coverprofile` 로 각 분기 본문 블록 실행 횟수 > 0 확인.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range at 317:2` | 줄이 있는 기록 | `TestProgressPointsAtTheRestartWhenOneIsPending` (`report_test.go:160`) | 해당 없음(기준선) | 본문 실행 확인 |
| B2 | `if at 318:3` | 첫 줄에서 계좌 참조 채택 | `TestProgressPointsAtTheRestartWhenOneIsPending` (`report_test.go:160`) | 해당 없음(기준선) | 본문 실행 확인 |
| B3 | `range at 322:2` | 카탈로그 전 단계 | `TestProgressOnAnUnstartedVerification` (`report_test.go:185`) | 해당 없음(기준선) | 본문 실행 확인 |
| B4 | `if at 324:3` | 줄 없는 단계 → Pending | `TestProgressOnAnUnstartedVerification` (`report_test.go:185`) | 해당 없음(기준선) | 본문 실행 확인 |
| B5 | `if at 331:3` | 존속 단계가 새 프로세스 대기 | `TestProgressPointsAtTheRestartWhenOneIsPending` (`report_test.go:160`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B6 | `range at 336:2` | M0 체크포인트 줄 순회 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` (`m0_status_test.go:9`) | 해당 없음(기준선) | 본문 실행 확인 |
| B7 | `if at 337:3` | M0 체크포인트가 status 에 보임 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` (`m0_status_test.go:9`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |

편집(task 3.x) 뒤 추가할 RED: 대사 줄이 있는 기록에서 (a) Steps·Pending·AwaitingRestart 불변, (b) 대사된 artifact 가
Outstanding 에 없고 "reconciled absent" 로만 표시.
