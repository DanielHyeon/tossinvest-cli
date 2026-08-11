# Function Logic Map: `Journal.MarkAlertDelivered`

- Source: `internal/journal/outbox.go` (337-348)
- AST evidence: `ast.json` — branches 1, returns 2, calls 5, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판 D0.3의 나머지 절반 —
*"두 경로가 같은 행을 동시에 정산해도 하나만 성공한다"* — 이 여기 UPDATE의
`WHERE` 절에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | `alert_outbox.id` | `ClaimAlertForDelivery`가 돌려준 것 | 없는 id면 `ErrAlertNotFound` |
| 행의 현재 `state` | **`PENDING`이어야 한다** | `alert_outbox` | 아니면 0행 갱신 → `ErrAlertNotFound` |
| `j.clk` | 주입 | 프로덕션 `clock.System()` | — |
| `ctx` | 호출자의 것 | 17판에서는 **배달 루프의 것** | 취소되면 B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:344` | `ExecContext` 실패 | 없음 | 오류 `:345` |
| — `:347` | — | 아래 표 | `requireOneRow(res, id)` |

**분기가 하나뿐인 것이 요점이다.** 판정은 Go의 `if`가 아니라 SQL의 `WHERE`에 있다:

```sql
WHERE id = ? AND state = ?   -- state = PENDING (:342-343)
```

**이것이 compare-and-swap이다.** 행이 이미 `DELIVERED`면 0행이 갱신되고
`requireOneRow`(`:465-474`)가 `ErrAlertNotFound`를 돌려준다 — 조용한 성공이 아니라
**오류**다. 그래서 두 경로가 같은 행을 정산하려 해도 **하나만 성공한다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.clk.Now` `:338` | `delivered_at`·`last_attempt_at` | 주입 시계 | AST calls |
| `RFC3339` `:338` | 문자열화 | 순수 | AST calls |
| `j.db.ExecContext` `:339` | CAS UPDATE | 로컬 SQLite — 밀리초 | AST calls |
| `requireOneRow` `:347` | **0행을 오류로 승격** | `ErrAlertNotFound` | AST calls |

**네트워크 없음. 잠금 없음**(Go 수준). 배제는 전부 SQLite가 한다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `state` → `DELIVERED` | `:341` | 내구. 이 행은 `PendingAlerts`에서 빠진다 |
| `delivered_at`·`last_attempt_at` | `:341` | 이번 시각 |
| `attempts` | `:341` | **`attempts + 1`** — 성공도 시도로 센다 |
| `last_error` | `:341` | `''`로 비움 |

- fallback 없음. 실패는 오류로 올라간다.
- **`attempts + 1`이 a096 계약의 계수다.** 17판이 "사이클당 행당 시도 1회"를
  SHALL로 쓰는 근거가 이 컬럼이고, 그래서 시도 수를 다시 셀 수 있다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 알림이 "배달됨"으로 넘어가는 유일한 자리다.
  17판이 `n.mu`를 기록 경로에서 떼는 근거의 절반이 여기 CAS이므로,
  **`WHERE ... AND state = ?`가 사라지면 D0.3은 무효다.**
- 17판에서 이 함수의 **호출자가 바뀐다**: 지금은 `deliver`/`Flush`가
  `n.mu`를 쥔 채 부르고, 17판에서는 배달 루프가 잠금 없이 부른다.
  **함수는 그대로이고 배제의 근거만 Go 뮤텍스에서 SQL 술어로 옮겨간다.**
  그 이동이 실제로 성립하는지는 R17-3이 관측한다.
- `ErrAlertNotFound`를 호출자가 어떻게 다루는지는 이 함수 밖이다.
  `Flush:455-457`은 그것을 **오류로 올려 루프를 끊는다** — 17판 구현이
  이 처리를 그대로 옮기면 이미 배달된 행 하나가 나머지 backlog를 막는다.
  §6.0 R17-6이 이 경계를 지고 간다.
