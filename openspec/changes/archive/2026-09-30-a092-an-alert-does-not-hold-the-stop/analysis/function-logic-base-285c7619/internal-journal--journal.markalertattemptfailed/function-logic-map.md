# Function Logic Map: `Journal.MarkAlertAttemptFailed`

- Source: `internal/journal/outbox.go` (352-363)
- AST evidence: `ast.json` — branches 1, returns 2, calls 5, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판이 *"nil publisher도 실패한
시도로 센다"*를 SHALL로 쓰는데, 그 "센다"의 구현이 이 함수다. HEAD의 `Flush`는
nil publisher에서 이 함수를 **부르지 않고 `break`한다**(`notifier.go:442-444`) —
그것이 17판이 고치는 것 중 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | `alert_outbox.id` | claim이 돌려준 것 | 없으면 `ErrAlertNotFound` |
| 행의 현재 `state` | **`PENDING`이어야 한다** | `alert_outbox` | 아니면 0행 → `ErrAlertNotFound` |
| `cause` | 임의 문자열 | 전송 오류의 `Error()` | 검증 없음 — 그대로 저장 |
| `ctx` | 호출자의 것 | 17판에서는 배달 루프의 것 | 취소되면 B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:359` | `ExecContext` 실패 | 없음 | 오류 `:360` |
| — `:362` | — | 아래 표 | `requireOneRow(res, id)` |

판정은 여기서도 SQL에 있다 — `WHERE id = ? AND state = ?`(PENDING, `:357-358`).
**행은 `PENDING`으로 남는다**: state를 바꾸지 않으므로 다음 주기의
`PendingAlerts`가 같은 행을 다시 집는다. 이것이 "critical 알림은 네트워크가
죽었다고 버려지지 않는다"의 구현이다(doc comment `:350-351`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.clk.Now` `:353` | `last_attempt_at` | 주입 시계 | AST calls |
| `RFC3339` `:353` | 문자열화 | 순수 | AST calls |
| `j.db.ExecContext` `:354` | CAS UPDATE | 로컬 SQLite | AST calls |
| `requireOneRow` `:362` | 0행을 오류로 | `ErrAlertNotFound` | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `attempts` | `:356` | **`attempts + 1`** |
| `last_attempt_at` | `:356` | 이번 시각 |
| `last_error` | `:356` | `cause` 그대로 |
| `state` | — | **바뀌지 않는다.** `PENDING` 유지 |

- fallback 없음.
- **재무장이 `attempts`를 0으로 되돌린다**(`ClaimAlertForDelivery:232`).
  그래서 `attempts`는 "이 조건에 대한 총 시도"가 아니라 **"이번 에피소드의 시도"**다.
  a097이 정한 의미이고, 17판이 "사이클당 행당 1시도"를 셀 때 쓰는 단위도 이것이다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 실패한 알림이 보존되는 자리다. 17판이 배달을
  루프 밖으로 옮겨도 **이 함수가 불리는 횟수만큼만 시도가 세어진다**.
  그러므로 "시도 3회"가 계약으로 되살아나려면 배달 루프가 **행당 정확히 한 번**
  이 함수 또는 `MarkAlertDelivered`를 불러야 한다. §6.0 R17-6이 그것을 관측한다.
- **17판이 고치는 결함이 여기서 보인다**: HEAD `Flush`는 `n.Publisher == nil`일 때
  `break`하므로 이 함수가 **한 번도 불리지 않는다** — 시도도 안 세어지고
  `last_error`도 비어 있다. 운영자가 outbox를 읽으면 "아무도 시도하지 않았다"와
  "시도했는데 전송기가 없었다"가 구별되지 않는다. §6.0 R17-4가 RED로 잡는다.
- 이 함수 자체는 그 결함의 원인이 아니다. **부르지 않은 것이 원인이다.**
