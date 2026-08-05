# Branch Test Map: `symbolsInDispute`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/reconcile/mismatch.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (437) `range` — for _, mismatch := range diff.Quantities { | `TestACreditIsSpentByTheFirstComparisonThatCouldUseIt`, `TestACreditWithNothingLeftToAnswerIsDiscarded`, `TestAReclassifiedDisagreementDoesNotReleaseTheBlock`, `TestAReclassifiedDisagreementRefutesTheCredit`, `TestARefutedCreditIsDiscardedPerSymbol`, `TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit`, `TestAnUnrelatedMismatchStillDoesNotDiscardAnAnsweredCredit` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B2 | (440) `range` — for _, missing := range diff.MissingOrders { | `TestACreditIsSpentByTheFirstComparisonThatCouldUseIt`, `TestACreditWithNothingLeftToAnswerIsDiscarded`, `TestAReclassifiedDisagreementDoesNotReleaseTheBlock`, `TestAReclassifiedDisagreementRefutesTheCredit`, `TestARefutedCreditIsDiscardedPerSymbol`, `TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit`, `TestAnUnrelatedMismatchStillDoesNotDiscardAnAnsweredCredit` | 측정: 아래 테스트가 이 줄을 실행한다 |
| B3 | (449) `range` — for _, external := range diff.ExternalPos { | `TestACreditIsSpentByTheFirstComparisonThatCouldUseIt`, `TestACreditWithNothingLeftToAnswerIsDiscarded`, `TestAReclassifiedDisagreementDoesNotReleaseTheBlock`, `TestAReclassifiedDisagreementRefutesTheCredit`, `TestARefutedCreditIsDiscardedPerSymbol`, `TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit`, `TestAnUnrelatedMismatchStillDoesNotDiscardAnAnsweredCredit` | 측정: 아래 테스트가 이 줄을 실행한다 |
