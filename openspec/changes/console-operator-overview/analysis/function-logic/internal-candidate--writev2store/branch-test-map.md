# Branch Test Map: `writeV2Store`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v2 파일을 연다 | 이 헬퍼 + `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows` | 아니오(사후 증거) | yes |
| B2 | v2 스키마를 그대로 만든다 | 같은 곳 | 아니오(사후 증거) | yes |
| B3 | 기준선을 이미 가진 후보 | 같은 곳 | 아니오(사후 증거) | yes |
| B4 | 창 안 12위와 1분 뒤 8위 | 같은 곳 | 기록됨(§4 P1) | yes |
| B5 | 순위 행 seed 실패 | 같은 곳 | 아니오(사후 증거) | yes |
| B6 | 1시간 뒤 승격된 두 번째 후보 | 같은 곳 | 기록됨(§4 P1 — D20) | yes |
| B7 | 승격 9분 **전**의 148위 행 — 삶 사이의 간극 | 같은 곳 | 기록됨(§4 P1 실행 재현) | yes |
