# Branch Test Map: `coverageAnswered`

- Source: `internal/candidate/scan.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 지지자를 하나씩 본다 | `TestOneSurvivingSupporterIsNotEnoughToCool` | 아니오(사후 증거) | yes |
| B2 | 설정에서 빠진 지지자는 요구하지 않는다 | `TestASupporterThatLeftThePanelDoesNotBlockCoolingForever`, `TestASourceTheSchedulePassedOverIsNotASourceThatIsGone` | 기록됨(§2-5) | yes |
| B3 | 들을 수 있었는데 답하지 않은 지지자가 있으면 false | `TestAScanDoesNotCoolASymbolItDidNotLookFor`, `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` | 기록됨(§5 리뷰 P1-1 실행 재현) | yes |
