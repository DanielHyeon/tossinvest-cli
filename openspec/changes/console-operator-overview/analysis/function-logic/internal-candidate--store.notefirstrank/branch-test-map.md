# Branch Test Map: `Store.NoteFirstRank`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B6·B9·B11·B13·B14는 직접 테스트가 없다 — SQLite I/O와 손상된 저장 값 경로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 인자 검증 진입 | `TestARankOfZeroIsNotAFirstSighting` | 아니오(사후 증거) | yes |
| B2 | rank 0·total 0·음수는 위치가 아니다 | `TestARankOfZeroIsNotAFirstSighting` | 기록됨(§1 7절 — rank/0 = +Inf) | yes |
| B3 | 목록 길이보다 큰 순위 | `TestARankOfZeroIsNotAFirstSighting` | 기록됨(§4 P2 — `PercentileExceeds`가 `Rank > RankTotal`을 막지 않았다) | yes |
| B4 | 순간 없는 최초 목격 | `TestARankOfZeroIsNotAFirstSighting` | 아니오(사후 증거) | yes |
| B5 | 원천 없는 최초 목격 | `TestARankOfZeroIsNotAFirstSighting` | 기록됨(§3 P2 — trim 안 된 source가 WARMING_UP이 된다) | yes |
| B6 | 트랜잭션 시작 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B7 | 후보 행 읽기 결과를 가른다 | `TestARankOfZeroIsNotAFirstSighting` | 아니오(사후 증거) | yes |
| B8 | 대상 없는 write는 `ErrNoCandidate` — 조용히 성공하지 않는다 | `TestARankOfZeroIsNotAFirstSighting` | 기록됨(§1 7절 — `NoteSources`가 대상 없이 조용히 성공) | yes |
| B9 | 후보 행 읽기 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B10 | 이미 기록된 순위는 덮이지 않는다 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`, `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank` | 기록됨(§4 P1 — D20 실행 재현) | yes |
| B11 | 읽을 수 없는 `first_seen_at` | 없음 | 아니오(사후 증거) | 미커버 |
| B12 | 정체성 창 밖 관측은 저장도 에러도 아니다 | `TestARankFromOutsideTheIdentityWindowIsNotStored` | 기록됨(§4 P1 — 만료 9분 전 148위 행이 새 삶의 최초 관측이 되던 것을 실제 저장소로 재현) | yes |
| B13 | write 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B14 | 커밋 실패 | 없음 | 아니오(사후 증거) | 미커버 |
