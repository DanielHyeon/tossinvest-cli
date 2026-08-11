# Function Logic Map: `Journal.MarkAlertDelivered`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 함수의 CAS가 배제로 오해된다.** `WHERE id = ? AND state = ?`는 진짜 CAS이고
> 정확히 하나만 성공시킨다. 그러나 **그것은 publish가 끝난 뒤에 돈다** —
> `claimAndDeliver` 이탈 `:276` → `deliver` → publish 성공 → `notifier.go:356`.
>
> 그래서 이 CAS가 막는 것은 **이중 정산**이고 이중 발송이 아니다.
> 두 번째 발송자의 푸시는 이 UPDATE가 실패하기 전에 이미 도착해 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 존재하는 행 | 호출자가 claim에서 받은 값 | 없거나 PENDING이 아니면 `requireOneRow`가 `ErrAlertNotFound` |
| 행의 `state` | **PENDING이어야 한다** | `:343`의 CAS 인자 | 아니면 0행 → 오류 |
| `j.clk.Now()` | 주입된 시계 | `:338` | — |
| **`token`** | **오늘 없다** | — | **a099가 더하는 인자다** — 이름이 아니라 **토큰**이다(design D3의 ABA). 반환도 `SettleOutcome`이 된다(C3) |

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 1 · 이탈 2 · 호출 5.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:344` | `ExecContext` 실패 | 없음 | 이탈 `:345` 오류 | 기존 |
| — 이탈 `:347` | 정상 | **UPDATE 하나** (`:340-343`) | `requireOneRow(res, id)` | 기존 + **a099 R13** |

**분기가 하나뿐이다.** 판정은 전부 SQL 술어 안에 있고 Go 쪽에는 오류 검사만 있다.
a099가 더하는 소유자 비교도 **같은 자리** — 술어 안이다. 분기는 안 는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RFC3339(j.clk.Now())` `:338` | `delivered_at`·`last_attempt_at` | — | `ast.json` calls |
| `j.db.ExecContext` `:339` | **CAS UPDATE** — `state`, `delivered_at`, `last_attempt_at`, `attempts+1`, `last_error=''` | 트랜잭션 없음. 단일 문장이 원자적이다 | 같음 |
| `fmt.Errorf` `:345` | 오류 포장 | — | 같음 |
| `requireOneRow` `:347` | 0행이면 `ErrAlertNotFound` | `outbox.go:465-474` — *"or it is no longer pending"* | 같음 |

**live bindings — 프로덕션 호출자 둘, 둘 다 `internal/obs`**:
`notifier.go:356` (`deliver` 안, publish 성공 뒤) · `notifier.go:455` (`Flush` 안).

## State mutations and fallbacks

- **UPDATE 하나**: PENDING → DELIVERED, `attempts`를 하나 올린다.
- **폴백 없다.** 0행이면 오류를 돌려주고 끝이다. 호출자가 그것을 어떻게 다루는지는
  이 함수 밖이다 — `notifier.go:356`은 로그로, `Flush` `:455-457`은 **반환으로**
  다룬다(루프를 끊는다).
- a099가 더하는 것: 같은 UPDATE에서 **임차 열을 비운다**. state가 DELIVERED가 되므로
  claim 가능 조건(`state='PENDING'`)에서 어차피 빠지지만, 재무장이 이 행을 다시
  PENDING으로 되돌릴 수 있으므로(`ClaimAlertForDelivery:229`) **남은 임차는 지운다.**

## Safety conclusion

- **Safe edit boundary**: a099는 **UPDATE의 SET에 「네 열 초기화」(design C5), WHERE에
  토큰 조건 하나**를 더하고 시그니처에 `token`을 더한다. B1의 조건과 두 이탈의 의미는
  안 바꾼다. **0행일 때 같은 트랜잭션에서 한 번 더 읽어 네 갈래를 가른다**(C3).
  편집 후 AST의 branches가 여전히 1이면 Go 쪽 제어 흐름 무변화다.
- **High-risk impact**: **yes** — 원장. 이 UPDATE가 실패하면 행이 PENDING으로 남고
  진입 게이트가 잠긴 채로 있다. **그 방향이 옳다** (열린 쪽으로 실패).
- **덮이지 않은 것을 이름으로 적는다**:
  - **소유자 비교를 더하면 실패 경로가 하나 는다** — 임차를 잃은 발송자의
    `MarkAlertDelivered`가 0행을 받는다. 그 발송자는 **이미 publish했다.**
    행은 PENDING으로 남고 다음 발송자가 또 보낸다. **at-least-once의 대가이고
    design D8이 그것을 명시한다.** R13이 이 경로를 관측한다.
  - `attempts`의 상한을 이 함수가 안 본다. 예산은 `deliver` 쪽이다 — a092의 표면.
