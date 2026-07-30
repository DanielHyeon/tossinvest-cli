# Branch Test Map: `TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | — (동작 무변경) | yes |
| B2 | 기준선을 기록한다 | 이 테스트 | — (동작 무변경) | yes |
| B3 | 냉각시킨다 | 이 테스트 | — (동작 무변경) | yes |
| B4 | 냉각 중에도 기준선이 남는다 | 이 테스트 | 기록됨(§3 P0-3) | yes |
| B5 | 재승격한다 | 이 테스트 | — (동작 무변경) | yes |
| B6 | 재진입이 `first_seen_at`을 옮기지 않는다 | 이 테스트 | 기록됨(§1 2절 — D1 우회) | yes |
| B7 | 재진입 뒤 기준선을 읽는다 | 이 테스트 | — (동작 무변경) | yes |
| B8 | 재진입이 새 기준선을 만들지 않는다 | 이 테스트 | 기록됨(§3 P0-3) | yes |
| B9 | 만료 시점 상태를 읽는다 | 이 테스트 | — (동작 무변경) | yes |
| B10 | staleness만으로 만료에 닿는다 | 이 테스트 | 기록됨(§1 2-a) | yes |
| B11 | 만료 후 다시 승격한다 | 이 테스트 | — (동작 무변경) | yes |
| B12 | 만료 후 `first_seen_at`은 새 pass다 | 이 테스트 | 기록됨(§1 2-c) | yes |
| B13 | 만료 후 기준선을 읽는다 | 이 테스트 | — (동작 무변경) | yes |
| B14 | 새 후보는 죽은 후보의 기준선을 상속하지 않는다 | 이 테스트 | 기록됨(§3 P0-3) | yes |
| B15 | 기준선의 순간·원천도 초기화된다 | 이 테스트 | — (동작 무변경) | yes |
| B16 | 새 삶이 자기 기준선을 쓴다 | 이 테스트 | — (동작 무변경) | yes |
| B17 | 그 값이 새 삶의 것이다 | 이 테스트 | — (동작 무변경) | yes |
