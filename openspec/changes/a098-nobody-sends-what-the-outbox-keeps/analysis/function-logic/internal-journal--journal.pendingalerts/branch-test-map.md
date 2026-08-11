# Branch Test Map: `Journal.PendingAlerts`

Source: `internal/journal/outbox.go` (392-405). AST 기준 branches 2 / returns 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:395` `limit > 0` → `LIMIT` 절. **두 갈래를 한 줄에 적는다** — 거짓 갈래만 덮여 있다 | 거짓(`limit<=0`): `TestReplayKeyConflictEnqueuesTheCriticalAlert` (`internal/execgw/replay_test.go:628`). **참(`limit>0`): 없음 — 트리 전체에 그런 호출이 0건** | no | **부분** |
| B2 | `:400` 질의 오류 | **없음** — 배수 중 원장을 깨는 테스트가 없다 | no | **no** |

## `LIMIT` 경로는 오늘 죽은 코드다

호출 지점 전수(프로덕션 2 + 테스트 7)에서 **`limit` 인자가 전부 `0`이다.**

```text
internal/obs/notifier.go:437          PendingAlerts(ctx, 0)   Flush
internal/obs/notifier.go:491          PendingAlerts(ctx, 0)   Acknowledge
internal/journal/outbox_test.go:49    PendingAlerts(ctx, 0)
internal/execgw/replay_test.go:638    PendingAlerts(..., 0)
internal/obs/obs_test.go:407·480·516·579   PendingAlerts(ctx, 0)
internal/obs/a096_one_send_per_condition_test.go:405   PendingAlerts(ctx, 0)
```

`limit > 0`인 호출은 **하나도 없다.** 그러므로 B1의 참 갈래는 프로덕션에서도
테스트에서도 실행된 적이 없다.

**이것이 a092에 주는 값이다** — a092의 배치 설계(`alertFlushBatch`)는 이 손잡이를
쓸 계획이고, 그 손잡이는 **한 번도 돌려진 적이 없다.** a098은 배치를 도입하지 않으므로
이 갈래를 켜지 않는다. **a092가 켤 때 RED가 필요하다는 사실을 여기 적어 둔다.**

## 필요한 RED

| # | Branch | Scenario | 기대 | 소유 |
|---|---|---|---|---|
| — | B1 참 | PENDING 5행에 `limit=2` | 2행, `id` 오름차순으로 앞의 둘 | **a092** (배치를 도입하는 change) |
| R2 | B2 | 배수 중 질의 오류 | a098의 루프가 죽지 않는다 | a098 |

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 2, returns 2)
- 호출 지점 전수: `rg 'PendingAlerts\(' -g '*.go'` → 정의 1 + 호출 9
- `attempts` 열이 실려 온다: `alertSelect` `outbox.go:387`, `Alert.Attempts` `:95`
