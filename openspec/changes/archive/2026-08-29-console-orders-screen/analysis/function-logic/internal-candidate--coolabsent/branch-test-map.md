# Branch Test Map: `coolAbsent`

- Source: `internal/candidate/scan.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보 읽기 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B2 | 저장된 후보를 순회한다 | `TestASymbolThatLeavesEveryListCools` | 아니오(사후 증거) | yes |
| B3 | 이번에 올라온 심볼·다른 시장·이미 냉각된 후보는 건너뛴다 | `TestASymbolThatLeavesEveryListCools` | 아니오(사후 증거) | yes |
| B4 | 지지자가 다 답하지 않았으면 냉각하지 않는다 | `TestAScanDoesNotCoolASymbolItDidNotLookFor`, `TestOneSurvivingSupporterIsNotEnoughToCool`, `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` | 기록됨(§1 파일 헤더의 429 시나리오, §5 리뷰 P1-1 실행 재현) | yes |
| B5 | 냉각 write 실패 | 없음 | 아니오(사후 증거) | 미커버 |
