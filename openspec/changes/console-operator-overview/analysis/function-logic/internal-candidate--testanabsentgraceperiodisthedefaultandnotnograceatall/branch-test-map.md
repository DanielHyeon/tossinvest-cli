# Branch Test Map: `TestAnAbsentGracePeriodIsTheDefaultAndNotNoGraceAtAll`

- Source: `internal/candidate/store_test.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보를 만든다 | 이 테스트 | 아니오(사후 증거) | yes |
| B2 | 즉시 냉각시킨다 | 이 테스트 | 아니오(사후 증거) | yes |
| B3 | grace 0과 음수를 둘 다 시도한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B4 | 스윕을 돌린다 | 이 테스트 | 아니오(사후 증거) | yes |
| B5 | 미설정 유예가 '유예 없음'이 되지 않는다 | 이 테스트 | 기록됨(§4 P0 — 임계 0이 전 목록을 통과, 같은 실패 형태) | yes |
| B6 | 기본값 창을 지나 다시 스윕한다 | 이 테스트 | 아니오(사후 증거) | yes |
| B7 | 기본값이 실제로 적용된 것이지 스윕이 죽은 것이 아니다 | 이 테스트 | 기록됨(§4 P2 — 가드를 지워도 green인 테스트 형태의 예방) | yes |
