# Branch Test Map: `releaseReJudgedQuarantineTx`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/journal/exit_snapshot.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (865) `if` — if reJudgingVersion <= 0 { | `TestAJudgementThatIsNotAReJudgementReleasesNothing`, `TestAReJudgementReleasesOnlyTheRowThatEarnedIt`, `TestAnOperatorOnlyQuarantineSurvivesAGenuineReJudgement` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B2 | (877) `if` — if err != nil \|\| !ok \|\| active.Reason != QuarantineReasonAmbiguousRecovery { | `TestAnOperatorOnlyQuarantineSurvivesAGenuineReJudgement` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B3 | (885) `if` — if active.Version != reJudgingVersion { | `TestAJudgementThatIsNotAReJudgementReleasesNothing`, `TestAReJudgementReleasesOnlyTheRowThatEarnedIt` | 측정: 아래 테스트가 이 줄을 실행한다 |
