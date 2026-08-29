# Branch Test Map: `TestARankOfZeroIsNotAFirstSighting`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 다섯 가지 잘못된 순위쌍을 전부 시도한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | rank 0·total 0·음수·역전은 전부 거부 | 이 테스트 | 기록됨(§1 7절 — rank_total 0이 +Inf로 게이트 통과, §4 P2 — Rank > RankTotal) | yes |
| B4 | 순간 없는 최초 목격은 거부 | 이 테스트 | 아니오(사후 증거) | yes |
| B5 | 원천 없는 최초 목격은 거부 | 이 테스트 | 기록됨(§3 P2) | yes |
| B6 | 대상 없는 provenance write는 `ErrNoCandidate` | 이 테스트 | 기록됨(§1 7절 — `NoteSources`의 조용한 성공) | yes |
| B7 | 모르는 후보 조회는 found=false | 이 테스트 | 아니오(사후 증거) | yes |
