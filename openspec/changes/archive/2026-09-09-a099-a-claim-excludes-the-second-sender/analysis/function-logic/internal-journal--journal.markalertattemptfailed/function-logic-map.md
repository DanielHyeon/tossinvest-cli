# Function Logic Map: `Journal.MarkAlertAttemptFailed`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **⚠⚠ 이 함수는 임차를 풀지 않는다.** a099의 초안은 여기서 푼다고 했고 **틀렸다.**
>
> `deliver`의 AST가 반증한다 — 그 번들은 `internal-obs--notifier.deliver`에 있다:
> 이 함수를 부르는 B8 `notifier.go:384` 다음이 B9 `:387`(예산이 남았나)와
> B10 `:388`(대기)이고, 그다음 루프가 `:354`로 **되돌아간다.**
>
> **실패 기록은 발송자가 끝났다는 뜻이 아니다.** 여기서 풀면 `:388`의 대기 동안
> 두 번째 발송자가 들어오고 원래 발송자는 임차 없이 또 보낸다.
>
> 이 함수가 하는 일은 **시도를 세는 것**뿐이다. 행이 PENDING으로 남는 것은 설계이고
> (`:350-351`의 주석 — *"a critical alert is not discarded because the network was
> down"*), a099는 그 의미를 안 바꾼다. 해제는 `ReleaseAlertClaim`이 한다 (design D3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 존재하는 PENDING 행 | 호출자가 claim에서 받은 값 | 0행이면 `ErrAlertNotFound` |
| `cause` | 임의 문자열 | 호출자 — `perr.Error()` (`notifier.go:452`) | 그대로 `last_error`에 들어간다 |
| 행의 `state` | **PENDING이어야 한다** | `:358`의 CAS 인자 | 아니면 0행 |
| **`token`** | **오늘 없다** | — | **a099가 더하는 인자다** — 이름이 아니라 **토큰**이다(design D3의 ABA). 반환도 `SettleOutcome`이 된다(C3) |

**불변식**: 이 함수는 **상태를 안 바꾼다.** `attempts`, `last_attempt_at`,
`last_error`만 쓴다 (`:355-358`). 행은 PENDING으로 남는 것이 설계다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 1 · 이탈 2 · 호출 5.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:359` | `ExecContext` 실패 | 없음 | 이탈 `:360` 오류 | 기존 |
| — 이탈 `:362` | 정상 | **UPDATE 하나** — `attempts+1`, `last_attempt_at`, `last_error` | `requireOneRow(res, id)` | 기존 + **a099 R6 · R13** |

`MarkAlertDelivered`와 **같은 모양이다** — 분기 1 / 이탈 2, 판정은 전부 술어 안.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RFC3339(j.clk.Now())` `:353` | `last_attempt_at` | — | `ast.json` calls |
| `j.db.ExecContext` `:354` | **CAS UPDATE** — 상태는 그대로 PENDING | 단일 문장, 트랜잭션 없음 | 같음 |
| `fmt.Errorf` `:360` | 오류 포장 | — | 같음 |
| `requireOneRow` `:362` | 0행이면 오류 | `outbox.go:465-474` | 같음 |

**live bindings — 프로덕션 호출자 둘, 둘 다 `internal/obs`**:
`notifier.go:384` (`deliver` 안, publish 실패 뒤 — 반환값을 검사하고 로그만 남긴다)
· `notifier.go:452` (`Flush` 안 — **반환값을 버린다**, `_ =`).

## State mutations and fallbacks

- **UPDATE 하나**, 상태 불변. `attempts`가 오르는 것이 예산 소진의 기록이다.
- **폴백 없다.** 0행이면 오류. `Flush`(`:452`)는 그것을 버리므로 이 함수의 실패는
  그 경로에서 **아무 데도 안 나타난다.**
- a099가 더하는 것: WHERE에 **소유자 조건 하나**뿐이다. **SET은 안 바뀐다** —
  임차를 그대로 둔다. 발송자가 아직 재시도할 것이기 때문이다 (위 ⚠⚠).
  임차를 잃은 발송자는 이 UPDATE가 0행이 되어 시도를 세지도 못한다 — R13.

## Safety conclusion

- **Safe edit boundary**: **WHERE에 토큰 조건 하나**와 시그니처의 `token`뿐이다.
  `MarkAlertDelivered`와 달리 **SET은 안 건드린다 — 임차를 유지한다**(design C5·D3의 ⚠⚠). B1과 두 이탈의 의미는 안 바꾼다.
  편집 후 branches가 여전히 1이면 제어 흐름 무변화다.
- **High-risk impact**: **yes** — 원장. 이 함수가 실패해도 행은 PENDING이고
  진입 게이트는 잠긴 채다. **열린 쪽으로 실패한다.**
- **덮이지 않은 것을 이름으로 적는다**:
  - **`Flush:452`가 반환을 버린다.** a099가 소유자 비교를 더하면 그 자리에서
    「임차를 잃었다」가 **조용히 사라진다.** a099는 그 호출을 `ClaimAlertByID`가
    성공한 행에 대해서만 하도록 바꾸므로 정상 경로에서는 안 생기지만,
    **만료가 그 사이에 일어나면 여전히 조용하다.** 이것은 a099가 남기는 알려진
    구멍이고 **contract 단계에서 `Flush`가 사라질 때 같이 사라진다.**
  - `attempts`의 상한 판정은 이 함수에 없다. `deliver` 쪽 — a092의 표면.
