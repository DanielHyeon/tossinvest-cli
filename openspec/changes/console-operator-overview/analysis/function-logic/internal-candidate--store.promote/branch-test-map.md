# Branch Test Map: `Store.Promote`

- Source: `internal/candidate/store.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B1·B2·B3·B4·B7·B8은 직접 테스트가 없다 — 인자 방어와 SQLite I/O 실패 경로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 시장·심볼 없는 승격 | 없음 | 아니오(사후 증거) | 미커버 |
| B2 | 순간 없는 승격 | 없음 | 아니오(사후 증거) | 미커버 |
| B3 | 트랜잭션 시작 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B4 | 기존 행 읽기 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B5 | 마지막 관측보다 이른 승격은 거부 | `TestAPromotionCannotRunBackwards`, `TestAnInstantFromThePastCannotExpireALiveCandidate` | 기록됨(§1 2-b 실행 재현) | yes |
| B6 | 냉각 중 재진입은 원래 `first_seen_at`을 지킨다 | `TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt`, `TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry`, `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | 기록됨(§4 P1 — D20 실행 재현) | yes |
| B7 | upsert 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B8 | 커밋 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B9 | 만료 후 승격은 죽은 삶의 출처·기준선·순위를 상속하지 않는다 | `TestANewCandidateDoesNotInheritTheDeadOnesSources`, `TestAReEntryAfterExpiryIsANewCandidate`, `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | 기록됨(§1 2-c 실행 재현) | yes |
