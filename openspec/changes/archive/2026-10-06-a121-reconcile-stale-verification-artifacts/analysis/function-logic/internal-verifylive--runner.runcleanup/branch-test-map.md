# Branch Test Map: `Runner.runCleanup`

- Source: `internal/verifylive/cleanup.go` (230-254)
- AST branches 8
- 측정: 고정 사본 `2c6ef1ef`(소스 sha 5cdcedb9… = base `de147cc2`)에서 `internal/verifylive` 시험 함수 전수를 하나씩
  `go test -run '^T$' -coverprofile` 로 돌려 분기 본문 블록 실행 횟수 > 0 확인. 편집 대상이 아니므로(근거 번들) RED/GREEN 칸은
  기준선만 적는다.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range at 232:2` | 대상 순회 | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` (`cleanup_test.go:182`) 외 4 | 해당 없음(비편집) | 본문 실행 확인 |
| B2 | `switch at 234:3` | 종류 분기 머리 | B3·B4 시험이 평가 | 해당 없음(비편집) | 머리 자체에 시작하는 블록 없음 — B3/B4 본문으로 확인 |
| B3 | `case at 235:3` | 일반 주문 취소 | `TestALeftoverOrderIsCancelledOnTheNextRun` (`cleanup_test.go:41`) 외 4 | 해당 없음(비편집) | 본문 실행 확인 |
| B4 | `case at 237:3` | 조건주문 취소(`CancelConditionalOrder`) | `TestAbortEndsAHeldChain` (`abort_test.go`) | 해당 없음(비편집) | 본문 실행 확인(이 시험만) |
| B5 | `case at 239:3` | 알 수 없는 종류 → continue | — | 해당 없음(비편집) | **미커버** — 현 작성기는 두 종류만 쓴다. a121 은 새 종류 값을 만들지 않으므로(Q2 (a)) 도달 경로 없음 |
| B6 | `if at 242:3` | 취소 성공 → continue | `TestALeftoverOrderIsCancelledOnTheNextRun` 외 3 | 해당 없음(비편집) | 본문 실행 확인 |
| B7 | `if at 245:3` | 첫 오류 보존 | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` | 해당 없음(비편집) | 본문 실행 확인(이 시험만) |
| B8 | `if at 248:3` | 계획 밖·취소 → break | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` | 해당 없음(비편집) | 본문 실행 확인(이 시험만) |

a121 RED 가 이 함수에 대해 고정하는 것은 "도달 불가" 다: `TestReconcileFilesCallNoWriteMethodAndAssertNoType`
(`internal/verifylive/reconcile_structure_test.go`) 와 `TestReconcileSucceedsWithOnlyGetRequestsAndTheTokenPost`
(`cmd/tossctl/verify_reconcile_seams_test.go`).
