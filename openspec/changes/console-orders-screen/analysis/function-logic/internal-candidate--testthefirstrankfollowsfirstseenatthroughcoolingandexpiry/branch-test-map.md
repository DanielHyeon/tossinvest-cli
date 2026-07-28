# Branch Test Map: `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 148/150을 최초 목격으로 기록 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | 냉각시킨다 | 이 테스트 | 아니오(사후 증거) | yes |
| B4 | 냉각 중에도 순위가 남는다 | 이 테스트 | 기록됨(§4 P1 — D20) | yes |
| B5 | 재승격한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B6 | 재진입이 `first_seen_at`을 옮기지 않는다 | 이 테스트 | 기록됨(§1 2절) | yes |
| B7 | 재진입의 순위 제안은 에러가 아니라 무시 | 이 테스트 | 아니오(사후 증거) | yes |
| B8 | 재진입 뒤 순위를 읽는다 | 이 테스트 | 아니오(사후 증거) | yes |
| B9 | 5위까지 오른 종목이 5위를 '처음 본 자리'로 만들지 못한다 | 이 테스트 | 기록됨(§4 P1 실행 재현) | yes |
| B10 | 만료 시점 상태를 읽는다 | 이 테스트 | 아니오(사후 증거) | yes |
| B11 | staleness만으로 만료에 닿는다 | 이 테스트 | 기록됨(§1 2-a) | yes |
| B12 | 만료 후 다시 승격한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B13 | 만료 후 순위를 읽는다 | 이 테스트 | 아니오(사후 증거) | yes |
| B14 | 새 후보가 죽은 후보의 순위를 상속하지 않는다 | 이 테스트 | 기록됨(§4 P1 — 죽은 삶이 산 삶을 대신 답한다) | yes |
| B15 | 순위의 순간·원천도 초기화된다 | 이 테스트 | 아니오(사후 증거) | yes |
| B16 | 새 삶이 자기 순위를 쓴다 | 이 테스트 | 아니오(사후 증거) | yes |
| B17 | 그 값이 실제로 돌아온 자리다 | 이 테스트 | 아니오(사후 증거) | yes |
| B18 | 저장된 순위가 `seen_late`까지 간다 — 컬럼이 채워진 것과는 다른 주장 | 이 테스트 | 기록됨(§4 P1) | yes |
