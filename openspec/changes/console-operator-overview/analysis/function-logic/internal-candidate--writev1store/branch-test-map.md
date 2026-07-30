# Branch Test Map: `writeV1Store`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v1 파일을 연다 | 이 헬퍼 + `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows` | 아니오(사후 증거) | yes |
| B2 | v1 스키마를 그대로 만든다 | 같은 곳 | 아니오(사후 증거) | yes |
| B3 | 가격 백필 대상 후보를 넣는다 | 같은 곳 | 아니오(사후 증거) | yes |
| B4 | 가격 행 두 건 | 같은 곳 | 아니오(사후 증거) | yes |
| B5 | 가격 행 seed 실패 | 같은 곳 | 아니오(사후 증거) | yes |
| B6 | 창 안 12위와 30분 뒤 4위 — v3 백필이 고를 것과 무시할 것 | 같은 곳 | 기록됨(§4 P1 — D20) | yes |
| B7 | 순위 행 seed 실패 | 같은 곳 | 아니오(사후 증거) | yes |
| B8 | 창 밖 순위밖에 없는 두 번째 후보 | 같은 곳 | 기록됨(§4 P1) | yes |
| B9 | 늦은 순위 행 seed 실패 | 같은 곳 | 아니오(사후 증거) | yes |
