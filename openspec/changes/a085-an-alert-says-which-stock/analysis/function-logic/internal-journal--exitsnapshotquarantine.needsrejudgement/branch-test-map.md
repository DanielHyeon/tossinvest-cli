# Branch Test Map: `ExitSnapshotQuarantine.NeedsReJudgement`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/journal/exit_snapshot.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | 분기 없음 — 단일 반환식의 happy path | `TestAJudgementThatIsNotAReJudgementReleasesNothing`, `TestANewQuarantineRecordsTheSelectorThatJudgedIt`, `TestAQuarantineFromTheCurrentSelectorIsNotReJudged`, `TestAQuarantineTheSelectorNeverDecidedIsNotReJudged`, `TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`, `TestAReJudgementReleasesOnlyTheRowThatEarnedIt`, `TestAReQuarantineDoesNotClaimAReJudgementThatDidNotHappen`, `TestARollbackDoesNotLetAnOlderSelectorOverturnANewerRefusal`, `TestAnOperatorOnlyQuarantineSurvivesAGenuineReJudgement`, `TestAnUnknownRevisionStillEarnsItsOneReJudgement`, `TestMigrationV30LeavesExistingQuarantinesUnstamped`, `TestTheOperatorReleasePathCannotStampMachineEvidence` | 측정: 분기가 없어 함수 진입이 곧 유일한 경로다 |
