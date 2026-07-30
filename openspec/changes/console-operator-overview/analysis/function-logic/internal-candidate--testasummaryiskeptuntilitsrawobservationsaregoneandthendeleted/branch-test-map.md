# Branch Test Map: `TestASummaryIsKeptUntilItsRawObservationsAreGoneAndThenDeleted`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 냉각시켜 삶의 끝을 정한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | 유예 1ns 전에 스윕한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B4 | 1ns 전에는 지우지 않는다 | 이 테스트 | 기록됨(§5 D16 — 보존 두 번째 층에 집행자 없음) | yes |
| B5 | 요약이 아직 있다 | 이 테스트 | 아니오(사후 증거) | yes |
| B6 | 유예 경계에서 스윕한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B7 | 경계에서 정확히 1건이 간다 | 이 테스트 | 기록됨(§5 D16) | yes |
| B8 | 그 요약이 실제로 사라졌다 | 이 테스트 | 아니오(사후 증거) | yes |
