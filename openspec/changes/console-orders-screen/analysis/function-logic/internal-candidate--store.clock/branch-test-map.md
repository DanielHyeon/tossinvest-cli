# Branch Test Map: `Store.Clock`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) watch 루프의 대기와 기록이 같은 시계에서 온다 | `TestTheWatchLoopWaitsOnTheInjectedClock` | 기록됨(§3 리뷰 P1-5 — 벽시계 철자 8종이 가드를 통과) | yes |
