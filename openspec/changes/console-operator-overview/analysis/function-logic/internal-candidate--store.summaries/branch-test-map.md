# Branch Test Map: `Store.Summaries`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B1·B3·B4·B5·B7·B8·B9·B10은 직접 테스트가 없다 — 손상된 저장 값과 SQLite I/O 실패 경로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 요약 질의 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B2 | 후보마다 상태와 두 저장된 사실을 함께 돌려준다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw`, `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank` | 기록됨(issues.md 4) | yes |
| B3 | 행 스캔 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B4 | 읽을 수 없는 `first_seen_at` | 없음 | 아니오(사후 증거) | 미커버 |
| B5 | 읽을 수 없는 `last_seen_at` | 없음 | 아니오(사후 증거) | 미커버 |
| B6 | 냉각된 후보의 `cooled_at`이 상태 유도에 들어간다 | `TestACycleSweepsTheSummariesItsOwnRetentionHasOrphaned` | 아니오(사후 증거) | yes |
| B7 | 읽을 수 없는 `cooled_at` | 없음 | 아니오(사후 증거) | 미커버 |
| B8 | 읽을 수 없는 기준선 순간 | 없음 | 아니오(사후 증거) | 미커버 |
| B9 | 읽을 수 없는 최초 목격 순간 | 없음 | 아니오(사후 증거) | 미커버 |
| B10 | 행 순회 오류 | 없음 | 아니오(사후 증거) | 미커버 |
