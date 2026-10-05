# Branch Test Map: `BuildReport`

- Source: `internal/verifylive/report.go` (169-220)
- AST branches 10
- 측정: base `de147cc2` 에서 시험별 `go test -run '^T$' -coverprofile` 로 각 분기 **본문 블록**(그 줄에서 시작하는 가장
  오른쪽 커버리지 블록)의 실행 횟수 > 0 을 확인했다. 아래 Test 칸은 그 측정에서 본문을 실행한 시험 중 하나다.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range at 180:2` | 줄이 하나 이상인 기록 | `TestUnverifiedValuesAreNotCountedAsAnswers` (`report_test.go:86`) | 해당 없음(기준선) | 본문 실행 확인 |
| B2 | `if at 181:3` | 첫 줄에서 계좌 참조 채택 | `TestUnverifiedValuesAreNotCountedAsAnswers` (`report_test.go:86`) | 해당 없음(기준선) | 본문 실행 확인 |
| B3 | `if at 184:3` | approval 줄이 Steps 에서 빠짐 | `TestAFullRunProducesAReportWithNoIdempotencyGaps` (`report_test.go:104`) | 해당 없음(기준선) | 본문 실행 확인 — `TestUnverifiedValuesAreNotCountedAsAnswers` 에선 미실행(스텝 줄뿐) |
| B4 | `range at 193:3` | 관측 두 개를 `latest` 에 | `TestReplayStaysDisabledUnlessBothHalvesArePositive` (`report_test.go:58`) | 해당 없음(기준선) | 본문 실행 확인 |
| B5 | `range at 197:2` | 빈 기록에서도 전 그룹 | `TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured` (`report_test.go:33`) | 해당 없음(기준선) | 본문 실행 확인 |
| B6 | `range at 199:3` | 빈 기록에서도 전 속성 | `TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured` (`report_test.go:33`) | 해당 없음(기준선) | 본문 실행 확인 |
| B7 | `if at 201:4` | 측정된 "false" 가 답으로 셈 | `TestUnverifiedValuesAreNotCountedAsAnswers` (`report_test.go:86`) | 해당 없음(기준선) | 본문 실행 확인 |
| B8 | `else at 203:11` | B7 거짓 → else-if 진입 | `TestUnverifiedValuesAreNotCountedAsAnswers` (`report_test.go:86`) | 해당 없음(기준선) | 본문 실행 확인 |
| B9 | `if at 203:11` | "unverified" 값은 Verified 아님 | `TestUnverifiedValuesAreNotCountedAsAnswers` (`report_test.go:86`) | 해당 없음(기준선) | 본문 실행 확인 |
| B10 | `if at 207:4` | 빈 기록 → 전 속성 Unverified | `TestBuildReportOnAnEmptyRecord` (`report_test.go:17`) | 해당 없음(기준선) | 본문 실행 확인 |

편집(task 3.x) 뒤 추가할 RED: 대사 줄이 있는 기록에서 (a) Steps·Unverified·ReplayEnabled 불변(B3 유지),
(b) 대사된 artifact 가 Outstanding 에 없고 새 표시 필드에 "reconciled absent" 로만 나옴.

## GREEN 로트 재추출 (a121 tasks 3.1·3.2, worktree a99a9059 + GREEN 편집)

- 편집: `rep.Reconciled = ReconciledArtifacts(entries)` 한 줄(분기 10 불변, 번호 불변 — difflib 정렬 equal, 줄만 이동). GREEN 관측: TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged·TestReconcileTextLabelsReconciledAbsentInReportAndStatus PASS; 변이 V-buildreport-reconciled.
- AST 재추출: `go run ./tools/logic-map` — 분기 10→10, 편집 전후 분기 열을 difflib 로 정렬한 결과 equal 뿐(재번호 없음), 앵커는 줄 이동만 반영.
