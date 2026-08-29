# Branch Test Map: `Schedule.UntilNextDue`

- Source: `internal/candidate/source.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 패널의 모든 원천 중 가장 이른 due를 고른다 | `TestARepeatedScanCannotBeAskedToPollFasterThanItsFloor` | 기록됨(§5 리뷰 P0) | yes |
| B2 | 한 번도 안 돈 원천은 지금 due | `TestTheWatchLoopWaitsOnTheInjectedClock` | 아니오(사후 증거) | yes |
| B3 | 이미 지난 원천은 0으로 보고한다 | `TestATickBelowTheSourceIntervalDoesNotEndTheDiscoveryLoop` | 기록됨(§5 리뷰 P0 — 아무것도 못 읽는 wake-up) | yes |
| B4 | 엔진 양보로 간격이 두 배가 되어도 루프가 끝나지 않는다 | `TestTheEngineYieldDoesNotEndTheDiscoveryLoop` | 기록됨(§5 리뷰 P0) | yes |
