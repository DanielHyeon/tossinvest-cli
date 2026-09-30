# Branch Test Map: `a066FailingNotices.EnqueueAlert`

- Source: `internal/app/engine/a066_risk_relaxation_test.go`(base `721d0338`). 삭제된 시험 가짜 — 분기 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | happy path(분기 없음, return at 215:2) | 상수 오류 반환 — 삭제 전 `TestA066ReleaseStandsWhenTheNoticeFails` 가 썼음, 이제 그 시험은 `a066FailingNotices.RecordCritical` 을 씀 | `TestA066ReleaseStandsWhenTheNoticeFails` | 해당 없음(시험 가짜) | 삭제됨 — 대체 경로 PASS |
