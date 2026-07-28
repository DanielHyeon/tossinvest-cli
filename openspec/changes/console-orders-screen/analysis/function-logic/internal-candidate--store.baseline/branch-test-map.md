# Branch Test Map: `Store.Baseline`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 저장된 기준선 세 컬럼을 읽는다 | `TestTheBaselineIsSetOnceAndNotOverwritten` | — (동작 무변경) | yes |
| B2 | 모르는 후보는 found=false — '기준선 없는 후보'와 다른 답이다 | `TestAnAbsentBaselineIsNotAZeroOne` | — (동작 무변경) | yes |
| B3 | 조회 실패 | 없음 | — (동작 무변경) | 미커버 |
