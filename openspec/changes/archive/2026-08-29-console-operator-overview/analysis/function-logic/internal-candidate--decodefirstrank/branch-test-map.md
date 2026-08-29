# Branch Test Map: `decodeFirstRank`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | NULL·비양수는 미기록이고 에러가 아니다 | `TestARankOfZeroIsNotAFirstSighting`, `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`(백필이 남긴 NULL) | 기록됨(§4 P1 — D20) | yes |
| B2 | 저장된 순간이 값과 함께 나온다 | `TestPruningRawObservationsLeavesTheFirstRankToo`, `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | 아니오(사후 증거) | yes |
| B3 | 읽을 수 없는 `first_rank_at` | 없음 | 아니오(사후 증거) | 미커버 |
