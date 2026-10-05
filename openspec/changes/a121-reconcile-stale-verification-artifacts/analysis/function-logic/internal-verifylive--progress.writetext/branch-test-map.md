# Branch Test Map: `Progress.WriteText`

- Source: `internal/verifylive/report.go` (346-379)
- AST branches 8
- 측정: 고정 사본 `2c6ef1ef`(소스 sha 4da7dc92… = base `de147cc2`)에서 `internal/verifylive` 시험 함수 전수를 하나씩
  `go test -run '^T$' -coverprofile` 로 돌려 분기 본문 블록 실행 횟수 > 0 확인.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 348:2` | 시작 전 기록 → 조기 반환 | `TestProgressOnAnUnstartedVerification` (`report_test.go`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B2 | `range at 356:2` | 단계 출력 | `TestProgressPointsAtTheRestartWhenOneIsPending` (`report_test.go`) | 해당 없음(기준선) | 본문 실행 확인 |
| B3 | `if at 359:2` | 남은 단계 | `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` | 해당 없음(기준선) | 본문 실행 확인 |
| B4 | `if at 362:2` | 재시작 대기 | `TestProgressPointsAtTheRestartWhenOneIsPending` | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B5 | `if at 366:2` | outstanding 머리줄·취소 권고 | `TestProgressPointsAtTheRestartWhenOneIsPending` | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B6 | `range at 368:3` | outstanding 행 | `TestProgressPointsAtTheRestartWhenOneIsPending` | 해당 없음(기준선) | 본문 실행 확인 |
| B7 | `if at 373:2` | M0 체크포인트 절 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` (`m0_status_test.go`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B8 | `range at 375:3` | 체크포인트 행 | `TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets` | 해당 없음(기준선) | 본문 실행 확인 |

편집(task 3.x) 뒤 GREEN 이어야 할 RED: `TestReconcileTextLabelsReconciledAbsentInReportAndStatus`
(`internal/verifylive/reconcile_projection_test.go`) — 대사 줄이 있는 기록에서 그 artifact 가 B5~B6 절(취소 권고 꼬리줄 포함)에
없고 "reconciled absent" 로 표시된다.
