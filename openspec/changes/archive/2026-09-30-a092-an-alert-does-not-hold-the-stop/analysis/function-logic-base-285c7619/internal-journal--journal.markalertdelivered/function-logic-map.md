# Function Logic Map: `Journal.MarkAlertDelivered`

- Source: `internal/journal/outbox.go` (337-348)
- AST evidence: `ast.json` — branches 1, returns 2, calls 5, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판 D0.3의 나머지 절반 —
*"두 경로가 같은 행을 동시에 정산해도 하나만 성공한다"* — 이 여기 UPDATE의
`WHERE` 절에 있다.

> ## ⛔ 2026-08-12 — a099가 이 문서의 절반을 무효화했다 (a099 §7.1)
>
> **아래 본문은 a092의 base commit 시점 코드를 기술한다.** 좌표도 AST도 그 시점의
> 것이고 **일부러 안 고쳤다** — 고치면 a092 자신의 `check_analysis`가 깨지고,
> 그때 남는 것은 「최신처럼 보이는 문서」다.
>
> **a099(`757550f1`)가 실제로 바꾼 것:**
>
> | 이 문서의 진술 | a099 이후 |
> |---|---|
> | `MarkAlertDelivered(ctx, id)` | **`MarkAlertDelivered(ctx, id, token) (SettleResult, error)`** |
> | 0행이면 `ErrAlertNotFound` **오류** | **네 결과 중 하나** — `SettleApplied`·`SettleLeaseLost`·`SettleAlreadySettled`·`SettleNotFound` |
> | `WHERE id = ? AND state = ?` | **`… AND claim_token = ?`가 붙는다** |
> | `attempts + 1`만 쓴다 | **임차 열 넷도 같이 비운다** |
>
> **그리고 이 문서의 핵심 주장 하나가 틀렸다.**
>
> *"두 경로가 같은 행을 정산하려 해도 하나만 성공한다"*는 참이지만,
> **그것이 「하나만 보낸다」를 뜻하지 않는다.** 이 CAS는 **publish가 끝난 뒤에 돈다** —
> `claimAndDeliver` → `deliver` → `Publish` 성공 → 이 함수. 두 번째 발송자의 푸시는
> 이 UPDATE가 실패하기 **전에** 이미 운영자 전화기에 도착해 있다.
> 2026-08-08의 `no such alert` 줄이 정확히 그 자리였다.
>
> **그러므로 「배제의 근거가 Go 뮤텍스에서 SQL 술어로 옮겨간다」는 이 함수만으로는
> 성립하지 않았다.** 성립시킨 것은 a099가 만든 **취득 시점의 임차**이고,
> 그 술어는 이 함수가 아니라 `alert_claim.go`의 `acquireAlertClaimTx`에 있다.
> 증거는 `openspec/changes/a099-…/analysis/function-logic/internal-journal--acquirealertclaimtx/`.
>
> **a092가 이 문서에서 인용해야 하는 것은 이제 그쪽이다.**
> 아래 「Safety conclusion」의 마지막 세 항목이 특히 그렇다 — `R17-3`이 재려던
> 「이동이 실제로 성립하는가」는 **a099 뒤에** 참이고, a099 없이는 거짓이다.

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
