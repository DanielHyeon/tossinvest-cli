# Branch Test Map: `Journal.PendingAlerts`

Source: `internal/journal/outbox.go` (392-405). AST 기준 분기 2 / 이탈 2 /
defers 1 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:395` **`limit > 0` → `LIMIT` 절이 붙는다** | **없음** — 프로덕션·테스트를 통틀어 이 함수를 0이 아닌 `limit`으로 부르는 자리가 없다 | no | no |
| B2 | `:400` `QueryContext` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| — `:404` | `limit = 0` → 미전달 행을 돌려준다 | `outbox_test.go TestEnqueueAlertIsIdempotentOnTheEventKey:49`, `internal/obs/obs_test.go TestPersistentDeliveryFailureBlocksEntries:409`·`TestCriticalAlertSurvivesAProcessRestart:578` | no | yes |
| — `:393` | **`ORDER BY id` — 오래된 것 먼저** | **없음.** 위 셋은 전부 `len(pending) != 1`로 단언한다 — **미전달 행이 둘 이상인 상태를 만드는 테스트가 없으므로 순서는 한 번도 관측되지 않았다** | no | no |

## a092가 이 함수에 대해 지는 것

**B1이 미테스트인 것이 이 표의 요점이다.** 17판은 배달 루프가 이 함수를
`alertFlushBatch = 8`로 부르게 하므로, **a092가 B1을 처음으로 실행시키는 change다.**

- **§6.0 R17-5**(사이클당 행 수 상한)가 B1을 RED로 처음 관측한다. 그 테스트는
  미전달 행을 상한보다 많이 만들어 두고 한 주기가 상한만큼만 처리하는지 본다.
  즉 **RED observed: 예정, GREEN observed: 예정** — 아직 어느 쪽도 관측되지
  않았으므로 위 표의 두 열은 `no`로 둔다. 관측한 뒤에 고친다.

B2는 SQLite 실패 주입이 필요하고 a092의 범위가 아니다
(`not-applicable`: 이 change는 이 함수를 편집하지 않는다).
