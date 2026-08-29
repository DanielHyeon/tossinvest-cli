# Branch Test Map: `TestARankFromOutsideTheIdentityWindowIsNotStored`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 창 밖 순위는 실패가 아니다 | 이 테스트 | 기록됨(§4 P1 — D20) | yes |
| B3 | 창 밖 순위는 반환값에서도 미기록 | 이 테스트 | 기록됨(§4 P1) | yes |
| B4 | 창 밖 순위는 저장소에도 없다 | 이 테스트 | 기록됨(§4 P1) | yes |
| B5 | 창 안 순위는 저장된다 | 이 테스트 | 아니오(사후 증거) | yes |
| B6 | 경계 1초 안쪽이 저장되어야 가드를 지운 구현이 통과하지 못한다 | 이 테스트 | 기록됨(§4 P2가 지적한 '맞는 이유로 통과하지 않는 테스트' 형태의 예방) | yes |
