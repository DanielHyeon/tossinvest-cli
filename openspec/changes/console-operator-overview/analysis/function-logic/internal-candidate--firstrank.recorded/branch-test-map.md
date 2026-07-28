# Branch Test Map: `FirstRank.Recorded`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 순위와 목록 길이가 **둘 다** 양수일 때만 기록된 것으로 읽는다 | `TestARankOfZeroIsNotAFirstSighting`, `TestARankWithoutItsListLengthIsRefused` | 기록됨(§1 7절 — `rank_total == 0`이 +Inf로 20%p 게이트를 통과) | yes |
