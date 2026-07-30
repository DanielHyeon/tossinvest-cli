# Branch Test Map: `Store.FreeSpace`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | prober 없는 저장소 | 없음 — `Open`이 항상 채운다 | 아니오(사후 증거) | 미커버 |
| B2 | probe 실패는 '넉넉함'이 아니라 정지다 | `TestSpaceItCouldNotMeasureIsNotSpaceItHas` | 기록됨(§5 D16) | yes |
| B3 | 잰 적 없는 잔여 공간은 0바이트가 아니다 | 없음 — 비-Linux `unsupportedProber`와 `Bsize<=0`의 프로덕션 경로 | 아니오(사후 증거) | 미커버 |
