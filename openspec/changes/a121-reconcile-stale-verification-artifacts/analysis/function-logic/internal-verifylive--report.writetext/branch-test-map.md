# Branch Test Map: `Report.WriteText`

- Source: `internal/verifylive/report.go` (239-291)
- AST branches 11
- 측정: 고정 사본 `2c6ef1ef`(= base `de147cc2` 와 같은 소스, sha 4da7dc92…)에서 `internal/verifylive` 의 시험 함수 전수를
  하나씩 `go test -run '^T$' -coverprofile` 로 돌려 각 분기 **본문 블록**(그 줄에서 시작하는 가장 오른쪽 블록)의 실행 횟수 > 0 을
  확인했다. B9~B11 은 패키지 안 시험이 하나도 실행하지 않아 `cmd/tossctl` 의 verify 시험 4개를
  `-coverpkg=./internal/verifylive` 로 따로 쟀다(실행: `TestVerifyReportKeepsReplayDisabledWithoutEvidence` 만).

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 245:2` | 빈 기록 → 조기 반환 | `TestBuildReportOnAnEmptyRecord` (`report_test.go`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B2 | `range at 251:2` | 단계 줄 출력 | `TestReportTextNamesTheUnverifiedProperties` (`report_test.go`) | 해당 없음(기준선) | 본문 실행 확인(이 시험만) |
| B3 | `range at 256:2` | 그룹 출력 | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B4 | `range at 258:3` | 속성 출력 | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B5 | `if at 260:4` | 빈 값 → `unverified` | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B6 | `if at 264:4` | 미검증 표지 `!` | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B7 | `if at 268:4` | 세부 줄 | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B8 | `range at 277:2` | 미검증 키 목록 | `TestReportTextNamesTheUnverifiedProperties` | 해당 없음(기준선) | 본문 실행 확인 |
| B9 | `if at 281:2` | outstanding 머리줄 | `TestVerifyReportKeepsReplayDisabledWithoutEvidence` (`cmd/tossctl/verify_test.go:502`) | 해당 없음(기준선) | 본문 실행 확인(cmd 측정, 패키지 안 0) |
| B10 | `range at 283:3` | outstanding 행 | 같음 | 해당 없음(기준선) | 본문 실행 확인(cmd 측정) |
| B11 | `if at 285:4` | deliberate 노트 | 같음 | 해당 없음(기준선) | 본문 실행 확인(cmd 측정) |

편집(task 3.x) 뒤 GREEN 이어야 할 RED: `TestReconcileTextLabelsReconciledAbsentInReportAndStatus`
(`internal/verifylive/reconcile_projection_test.go`) — 대사 줄이 있는 기록에서 (a) 그 artifact 가 B9~B10 절에 없고
(b) "reconciled absent" 라벨로 표시되며 (c) "취소"/"체결" 로 쓰이지 않는다.
