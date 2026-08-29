# Branch Test Map: `Store.Candidates`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 후보 질의 실패 | 없음 | — (동작 무변경) | 미커버 |
| B2 | 저장된 후보 전량을 상태와 함께 돌려준다 | `TestExpiryIsReadFromTheClockAndNotFromASweeper`, `TestTheHistoryOutlivesTheProcess` | — (동작 무변경) | yes |
| B3 | 읽을 수 없는 타임스탬프 | 없음 | — (동작 무변경) | 미커버 |
| B4 | 행 순회 오류 | 없음 | — (동작 무변경) | 미커버 |
