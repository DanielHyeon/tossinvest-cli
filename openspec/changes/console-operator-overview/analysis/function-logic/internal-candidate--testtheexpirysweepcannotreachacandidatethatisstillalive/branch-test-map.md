# Branch Test Map: `TestTheExpirySweepCannotReachACandidateThatIsStillAlive`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 세 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 승격 실패 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | active를 계속 본다 | 이 테스트 | 아니오(사후 증거) | yes |
| B4 | cooling을 방금 냉각시킨다 | 이 테스트 | 아니오(사후 증거) | yes |
| B5 | 스윕을 돌린다 | 이 테스트 | 아니오(사후 증거) | yes |
| B6 | 정확히 1건만 지워진다 | 이 테스트 | 기록됨(§5 D16) | yes |
| B7 | 살아 있어야 할 둘을 확인한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B8 | 조회 실패 | 이 테스트 | 아니오(사후 증거) | yes |
| B9 | 살아 있는 후보는 스윕이 닿지 못한다 | 이 테스트 | 기록됨(§1 2-a — 암묵 냉각이 없던 시절의 반대 방향) | yes |
| B10 | 암묵 냉각한 후보에는 닿는다 — 만료 순간이 어느 컬럼에도 없는데도 | 이 테스트 | 기록됨(§1 2-a 실행 재현) | yes |
