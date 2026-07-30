# Branch Test Map: `TestPruningRawObservationsLeavesTheFirstRankToo`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 순위·가격을 나르는 관측을 기록 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 승격 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | 최초 순위 기록 | 이 테스트 | 아니오(사후 증거) | yes |
| B4 | 보존 창 밖으로 prune | 이 테스트 | 아니오(사후 증거) | yes |
| B5 | prune 뒤 관측을 읽는다 | 이 테스트 | 아니오(사후 증거) | yes |
| B6 | 원 행이 정말 사라졌음을 먼저 확인한다 | 이 테스트 | 기록됨(§4 P2 — 가드를 지워도 green인 테스트 형태) | yes |
| B7 | 최초 순위를 읽는다 | 이 테스트 | 아니오(사후 증거) | yes |
| B8 | 순위·목록 길이·순간이 prune을 넘어 남는다 | 이 테스트 | 기록됨(§3 P0-3의 순위판 — D11 + D20) | yes |
| B9 | 원 행이 없어도 `seen_late`가 측정된다 | 이 테스트 | 기록됨(§4 P1 — 컬럼 이전에는 가장 오래 달린 후보에서 NO_FIRST_SIGHTING) | yes |
