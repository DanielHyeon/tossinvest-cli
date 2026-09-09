# Function Logic Map: `Journal.AcknowledgeAlert`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099는 이 함수를 편집하지 않는다.** 이 산출물이 있는 이유는 design D1과 D10이
> **이 함수의 술어를 근거로 쓰기 때문**이다. D1은 *"임차를 상태로 만들면
> `Acknowledge`가 0을 보고 게이트를 푼다"*를 주장하고, 그 주장은 이 함수가
> **무엇을 술어로 쓰는지**에 달려 있다.
>
> 1라운드 A-P9가 D1의 기전 서술이 틀렸다고 지적했고, 정정된 기전이 이 경로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 존재하는 행 | 호출자 | `requireOneRow`가 거절 (`:383`) |
| `operator` | **공백 아님** | 이 함수 (`:371`) | B1 → 이탈 `:372` 오류 |
| `alert_outbox.state` | **PENDING이어야 한다** | SQL 술어 `:379` | 아니면 0행 → `requireOneRow` 오류 |
| `j.clk.Now()` | 주입된 시계 | `journal.Journal` | — |

**불변식 — 이 함수가 지키는 것**: *"운영자의 해제는 이름을 남긴다."*
doc comment `:365-368`이 이유를 적는다 — 기계가 증명할 수 없던 것을 사람이 단언하는
자리이고 **audit trail이 요점이다**(불변식 5의 「인증」 축).

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 2 · 이탈 3 · 호출 8 · 대입 2.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:371` | `strings.TrimSpace(operator) == ""` | **없다 — 원장을 안 건드린다** | 이탈 `:372` 오류 | 기존 (`outbox_test.go:145`) |
| B2 `:380` | `ExecContext` 오류 | 시도했으나 실패 | 이탈 `:381` 포장된 오류 | **없음 — DB 오류 주입 없음** |
| — 이탈 `:383` | 정상 | **UPDATE 1행** (아래) | `requireOneRow(res, id)` | 기존 |

**분기 둘 다 안전 판정이 아니다.** 하나는 인자 검증, 하나는 오류 전파다.
**실제 판정 전체가 SQL 술어에 있다**: `WHERE id = ? AND state = ?`(`:378-379`,
인자 `AlertPending`).

### ⛔ 이 술어가 임차를 안 본다 — a099가 만드는 새 상호작용

a099가 임차 열을 더해도 이 UPDATE의 `WHERE`는 그대로다. 그러므로:

| 단계 | 무슨 일 |
|---|---|
| 1 | 발송자가 행을 claim하고 전송 중이다 (행은 PENDING) |
| 2 | 운영자가 **같은 행을 acknowledge한다** — 술어가 PENDING이므로 **성공한다** |
| 3 | 행이 ACKNOWLEDGED가 된다. **임차 열은 그대로 남는다** |
| 4 | 발송자의 전송이 성공한다 |
| 5 | `MarkAlertDelivered`도 `WHERE state = PENDING`이므로 **0행 → 오류** |
| 6 | D3의 규칙 *"보냈는지 원장이 모르면 임차를 놓지 않는다"*에 걸린다 |
| 7 | **ACKNOWLEDGED 행에 살아 있는 임차가 남는다** |

7의 결과 자체는 오늘 무해하다 — ACKNOWLEDGED 행은 `PendingAlerts`·`UndeliveredCount`
어디에도 안 나온다. **문제가 되는 것은 재무장이다.** `ClaimAlertForDelivery`의
재무장 UPDATE(`:229-236`)가 임차 열을 안 지우면, **재무장된 행이 이전 episode의
임차를 달고 태어난다.**

`:198-201`의 주석이 그 원칙을 이미 적는다 —
*"re-arming is the statement that this row is now a different episode"* ·
*"A row is evidence only when every column points at one event."*
**a097이 정한 원칙이고, 임차 열은 그 원칙의 예외가 아니다.**

tasks 4.6이 이것을 지목한다. **이 번들이 그 task의 근거다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:371`·`:379` | 운영자 이름 정규화 | — | `ast.json` calls |
| `errors.New` `:372` | 빈 이름 거절 | — | 같음 |
| `RFC3339(j.clk.Now())` `:374` | 승인 시각 | **주입된 시계** | 같음 |
| `j.db.ExecContext` `:375` | UPDATE 하나 | **트랜잭션 없음 — 단일 문** | 같음 |
| `fmt.Errorf` `:381` | 오류 포장 | — | 같음 |
| `requireOneRow` `:383` | 0행이면 오류 | 술어 불일치를 오류로 바꾼다 | 같음 |

**live bindings — 프로덕션 호출자 하나**:

```
internal/obs/notifier.go:500   if err := n.Journal.AcknowledgeAlert(ctx, id, operator); err != nil &&
```

그 호출자가 곧이어 `:506`에서 `UndeliveredCount`를 읽고 **0이면 진입 게이트를 푼다.**
**이 함수는 진입 게이트가 풀리는 경로 위에 있다.**

테스트 호출자(실측): `internal/journal/outbox_test.go:118,145` ·
`internal/journal/a096b_round2_test.go:87` ·
`internal/journal/a096_claim_for_delivery_test.go:141`.

## State mutations and fallbacks

- **mutation 하나**: `state = ACKNOWLEDGED` · `acknowledged_at` · `acknowledged_by`.
  세 열을 한 문장으로 쓴다(`:376-379`).
- **트랜잭션이 없다.** 단일 UPDATE라 원자성이 문장 자체에서 나온다.
  **a099가 임차 해제를 여기 붙이면 그 원자성 안에 들어가야 한다** — 별도 문장으로
  나누면 그 사이에 창이 생긴다.
- **폴백 없다.** 술어가 안 맞으면 `requireOneRow`가 오류다.
- **B1은 원장에 도달하기 전에 막는다** — 이름 없는 승인은 쓰기 자체가 없다.

## Safety conclusion

- **Safe edit boundary**: **없다 — a099는 이 함수를 편집하지 않는다.**
  편집 후 `source_sha256`이 그대로여야 한다(§5.3의 확인 대상).
  **만약 §4에서 임차 해제를 여기 붙이기로 바꾸면 그것은 Pre-Edit 재선언이다.**
- **High-risk impact**: **yes** — 운영자 인증 + 진입 게이트 해제 경로(불변식 5).
- **덮이지 않은 것을 이름으로 적는다**:
  - **B2(DB 오류)는 어떤 테스트도 안 덮는다.** 오늘부터 그렇다.
    **`not-applicable`: 이 change는 B2를 근거로 아무것도 주장하지 않는다.**
  - **acknowledge와 발송이 겹치는 창**(위 표 2단계)은 오늘도 존재하고 a099가
    만든 것이 아니다. a099가 바꾸는 것은 **그 창이 임차 열을 남긴다**는 점이고,
    그 처리는 tasks 4.6(재무장 시 초기화)이다.
  - `notifier.go:500`의 `err != nil &&` 뒤 조건은 이 번들의 범위 밖이다.
    **`not-applicable`: `Acknowledge`의 오류 처리는 a099가 안 건드린다.**
