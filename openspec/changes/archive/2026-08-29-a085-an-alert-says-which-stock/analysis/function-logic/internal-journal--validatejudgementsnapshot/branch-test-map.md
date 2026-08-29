# Branch Test Map: `validateJudgementSnapshot`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/journal/exit_snapshot.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B4. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (278) `if` — if err := stored.validate(); err != nil { | `TestOrderableSnapshotMustArmItsExactProposal`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B2 | (282) `if` — if line.PositionID != positionID \|\| line.ObservationID != judgement.Provenance.ObservationID \|\| | `TestOrderableSnapshotMustArmItsExactProposal`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B3 | (290) `if` — if expected.Zero() { | `TestOrderableSnapshotMustArmItsExactProposal`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B4 | (291) `if` — if judgement.Proposal != nil \|\| judgement.ArmSuppressedReason != "" { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B5 | (296) `if` — if judgement.Proposal == nil { | `TestOrderableSnapshotMustArmItsExactProposal`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B6 | (297) `if` — if !knownArmSuppression(judgement.ArmSuppressedReason) { | `TestOrderableSnapshotMustArmItsExactProposal`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B7 | (302) `if` — if judgement.ArmSuppressedReason != "" { | `TestOrderableSnapshotMustArmItsExactProposal` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B8 | (305) `if` — if judgement.Proposal.Action != string(expected.Action) \|\| judgement.Proposal.Level != expected.Level { | `TestOrderableSnapshotMustArmItsExactProposal` | 측정: 아래 테스트가 이 줄을 실행한다 |
