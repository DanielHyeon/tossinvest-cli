# Branch Test Map: `ExitObserver.workingSet`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/exitloop.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B7, B8, B9, B10, B12, B13, B16, B21, B22. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (492) `if` — if err != nil { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B2 | (496) `if` — if err != nil { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B3 | (500) `range` — for _, result := range stateResults { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B4 | (505) `range` — for _, p := range positions { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B5 | (506) `if` — if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B6 | (509) `if` — if !p.ExitEligible() { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B7 | (522) `if` — if !ok { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B8 | (524) `if` — if err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B9 | (525) `if` — if cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B10 | (530) `if` — if opened.PositionID == "" { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B11 | (539) `if` — if result.Corruption != nil { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B12 | (542) `if` — if qerr != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B13 | (543) `if` — if cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B14 | (553) `if` — if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B15 | (558) `else` — } else if active && !q.NeedsReJudgement() { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B16 | (554) `if` — if cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B17 | (558) `if` — } else if active && !q.NeedsReJudgement() { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B18 | (562) `else` — } else if active { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B19 | (562) `if` — } else if active { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B20 | (586) `if` — if identityErr != nil { | `TestAQuarantineThisSelectorWroteIsStillSkipped`, `TestAReJudgementDoesNotCancelAWorkingOrderBeforeItIsAllowedTo`, `TestAReJudgementNeverWithholdsAStop`, `TestASupersededQuarantineIsReJudgedAndReleased`, `TestASuppressedReJudgeArmingIsNotedAsADelay`, `TestTheReJudgementRetryIsSpentByTheAttempt` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B21 | (589) `if` — if qerr != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B22 | (590) `if` — if cycle.Err == nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
