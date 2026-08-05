# Branch Test Map: `validateExitEventArmSuppression`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/journal/exit_snapshot.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B1, B2, B10. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (454) `if` — if saved == nil && recomputed == nil && effective == nil && source == "" { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B2 | (455) `if` — if reason != "" { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B3 | (461) `if` — if recomputed == nil \|\| effective == nil { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B4 | (464) `if` — if recomputed.Line.PositionID != event.PositionID \|\| effective.Line.PositionID != event.PositionID \|\| | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B5 | (469) `switch` — switch source { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B6 | (470) `case` — case EffectiveSourceRecomputed: | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B7 | (471) `if` — if !reflect.DeepEqual(*effective, *recomputed) { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B8 | (476) `case` — case EffectiveSourceSaved: | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B9 | (477) `if` — if saved == nil \|\| !reflect.DeepEqual(*effective, *saved) { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B10 | (482) `case` — default: | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B11 | (486) `if` — if saved != nil { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B12 | (491) `if` — if err != nil \|\| selectedSource != expectedSource \|\| selected != effective.Line { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B13 | (495) `if` — if reason != "" && !knownArmSuppression(reason) { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B14 | (499) `if` — if (event.Action == "") != (event.ProposedIntentID == "") { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B15 | (502) `if` — if source == EffectiveSourceSaved { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B16 | (503) `if` — if armed \|\| reason != "" { | `TestSavedMonotoneRecoveryCannotArmRecomputedOrder` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B17 | (508) `if` — if recomputed.Line.Orderable && armed { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B18 | (509) `if` — if reason != "" \|\| event.Action != string(recomputed.Line.Action) { | `TestExitEventReadRejectsForgedArmSuppressionEvidence` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B19 | (515) `if` — if recomputed.Line.Orderable && !armed && !knownArmSuppression(reason) { | `TestExitEventReadRejectsForgedArmSuppressionEvidence`, `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B20 | (519) `if` — if !recomputed.Line.Orderable && (armed \|\| reason != "") { | `TestSavedMonotoneRecoveryCannotArmRecomputedOrder`, `TestTypedArmSuppressionPersistsOrderableSnapshotWithoutArming` | 측정: 아래 테스트가 이 줄을 실행한다 |
