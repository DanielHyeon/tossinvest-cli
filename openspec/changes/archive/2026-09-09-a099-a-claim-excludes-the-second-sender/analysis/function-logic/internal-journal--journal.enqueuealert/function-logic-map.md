# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 번들이 존재하는 이유는 design D13이다.** D13은 *"기록만 하는 호출자는 임차를
> 잡으면 안 된다"*를 주장하고, 그 주장은 **이 함수가 안에서 무엇을 하는지**를 근거로 쓴다.
> **함수 내부를 근거로 쓰는 문서는 AST 산출물이 먼저다.**
>
> 1라운드 A-P5가 이 경로를 지목했고, 1판 설계는 이 경로를 다루지 않았다.
>
> **§5.6 갱신(구현 후).** proposal 시점 이 함수는 분기가 **0**이었다 — 본문이
> `j.ClaimAlertForDelivery`로의 위임 한 줄이었다. 그 판이 이 번들의 발견이었고,
> 발견은 *"D13을 이 함수 안에서는 구현할 수 없다"*였다. **§4.7이 그 결론을 뒤집었다**:
> 위임을 끊고 이 함수에 자기 트랜잭션을 줬다. 지금 분기는 넷이다.
> **아래 표는 현재 HEAD의 열거이고, 위 문단은 그것이 어디서 왔는지의 기록이다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | 호출자 | `BeginTx`가 오류로 돌려준다 (B2) |
| `a Alert` | `EventKey`·`Type` 비어 있지 않음 | `alertKey` `:132` | B1이 트랜잭션을 열기 **전에** 거절한다 |
| `remindAfter` | **인자가 아니다 — `0` 리터럴** (`:146`) | **이 함수 자신** | — |
| `claimant` | **인자가 아니다 — 이 함수는 임차를 안 잡는다** | 본문에 `acquireAlertClaimTx` 호출이 없다 | — |

**불변식 — a099가 지켜야 하는 것**: *"기록만 하는 호출자는 발송 권한을 얻지 않는다."*

proposal 시점 이 불변식은 무의미했다 — 발송 권한이라는 것이 원장에 없었기 때문이다.
schemaV31이 그것을 만들었고, 그 순간 이 불변식이 실질을 갖는다.
**지금 이 함수가 그것을 지키는 방식은 「안 잡는 모드로 위임」이 아니라 「임차 코드가 없음」이다.**

## Branches and early returns

AST 열거 — 분기 4 · 이탈 5 · 호출 7 · 대입 4 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:133` | `alertKey(a)` 실패 | **없음 — 트랜잭션 전이다** | 이탈 `:134` `0, err` | `TestEnqueueAlertRequiresAKeyAndAType` |
| B2 `:137` | `BeginTx` 실패 | 없음 | 이탈 `:138` 포장 오류 | 없음 (드라이버 오류 미주입) |
| B3 `:147` | `recordAlertTx` 실패 | 롤백 (`defer :140`) | 이탈 `:148` `0, err` | 없음 |
| B4 `:150` | `tx.Commit()` 실패 | 롤백 | 이탈 `:151` 포장 오류 | 없음 |
| — 이탈 `:153` | 정상 | **행 하나 삽입 또는 재무장** (`recordAlertTx` 안) | `id, nil` | `TestEnqueueAlertIsIdempotentOnTheEventKey` · `TestARecordedAlertCanBeClaimedImmediately` |

### 이 함수가 자기 트랜잭션을 갖게 된 이유

proposal 판의 위임 형태(`return j.ClaimAlertForDelivery(...)`)를 유지하면
**임차 여부라는 판정이 인자로 들어간다.** 그러면 `ClaimAlertForDelivery`는
「청구한다」와 「청구하지 않는다」 두 모드를 갖고, 그 갈림이 실행 시점 인자에 달린다.

| 형태 | D13을 어디서 지키나 | 문제 |
|---|---|---|
| 위임 + 모드 인자 | `ClaimAlertForDelivery` 안의 새 분기 | 호출자가 인자를 틀리면 기록 전용 경로가 임차를 잡는다 |
| **자기 트랜잭션** (§4.7) | **호출 자체가 없다** | 인자로 틀릴 자리가 없다 |

`recordAlertTx`가 공통 절반이므로 중복은 SQL이 아니라 **트랜잭션 껍데기 12줄**이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `alertKey` `:132` | 입력 검증 + 중복 제거 키 | 트랜잭션 **전** — write lock을 잡고 거절하지 않는다 | `ast.json` calls |
| `j.db.BeginTx` `:136` | 쓰기 트랜잭션 | `_txlock=immediate`가 write lock을 즉시 잡는다. busy timeout은 `Options.BusyTimeout` | 같음 |
| `tx.Rollback` `:140` (defer) | 모든 이른 이탈에서 되돌린다 | 커밋 뒤 호출은 무해(no-op) | `ast.json` defers |
| `j.recordAlertTx` `:146` | **기록 절반** — dedup·삽입·재무장·`owed` 판정 | `remindAfter=0`을 넘긴다 | `ast.json` calls |
| `tx.Commit` `:150` | 확정 | 실패하면 defer가 롤백 | 같음 |
| `fmt.Errorf` `:138`, `:151` | 오류 포장 (`%w`) | — | 같음 |

**`recordAlertTx`의 두 번째 반환값을 `_`로 버린다** (`:146`). 그 값이 `owed`이고,
**이 함수가 그것을 안 보는 것이 D13 그 자체다** — 보낼지 말지를 판단하지 않으므로
청구할 이유가 없다.

**live bindings — 프로덕션 호출자 하나뿐이다**:

```
internal/execgw/replay.go:551   _, _ = g.journal.EnqueueAlert(ctx, journal.Alert{
```

**반환 둘을 다 버린다** (`_, _ =`). 이 호출자는 id도 오류도 안 본다.
**그러므로 이 호출자는 발송을 몰고 갈 수 없고, 몰고 갈 의도도 없다.**

테스트 호출자: `internal/journal/outbox_test.go` ·
`internal/journal/a096_claim_for_delivery_test.go` ·
`internal/journal/a099_lease_lifecycle_test.go` ·
`internal/journal/a099_regression_pins_test.go` ·
`internal/journal/a099_claim_excludes_the_second_sender_test.go` ·
`internal/obs/a099_lease_events_test.go`.

## State mutations and fallbacks

- **행 하나**: `recordAlertTx`가 삽입하거나 재무장한다. 이 함수 자신은 SQL을 안 쓴다.
- **`remindAfter = 0`을 넘긴다** (`:146`). doc comment `:142-145`가 이유를 적는다 —
  이 호출자는 발송하지 않으므로 알림 정책이 없고, 남의 몫으로 정산된 행을 재무장하면 안 된다.
  **다만 `recordAlertTx`는 인식 못 하는 state를 PENDING으로 복구할 수 있다.**
  그것은 시간 기반 재무장이 아니라 fail-safe 상태 복구이고, `remindAfter=0`이 막지 않는다.
- **임차 열 넷을 안 건드린다.** 삽입되는 행의 `claim_token`은 schemaV31의 기본값 `''`,
  `claim_expires_at`은 `NULL`이다 — 즉 **곧바로 청구 가능한 상태**로 태어난다.
- **폴백 없다.**

### D13이 없었으면 생겼을 일

| 단계 | 무슨 일 |
|---|---|
| 1 | `replay.go:551`이 critical 알림 행을 쓴다 |
| 2 | 그 호출이 임차를 잡는다 |
| 3 | 이 호출자는 **발송하지 않는다** — 반환을 버린다 |
| 4 | 그러므로 **아무도 그 임차를 안 푼다** |
| 5 | 배달 루프가 그 행을 **만료(81초)까지 못 집는다** |

**갓 기록된 critical 알림이 자기 임차 뒤에 갇힌다.**
`TestARecordedAlertCanBeClaimedImmediately`가 이 경로를 관측한다 —
`EnqueueAlert` 직후의 `ClaimAlertForDelivery`가 `ClaimAcquired`를 받아야 한다.

## Safety conclusion

- **Safe edit boundary**: 본문에 임차 호출이 없다는 것이 이 함수의 계약이다.
  `acquireAlertClaimTx`를 여기에 부르는 편집은 D13 위반이고, 위 표의 5단계를 되살린다.
- **High-risk impact**: **yes** — 원장 쓰기 경로이고(불변식 5),
  이 경로로 기록된 critical 알림이 갇히면 운영자가 못 받는다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B2·B3·B4에 테스트가 없다.** 드라이버·커밋 실패 주입이 이 패키지에 없다.
    a099가 만든 공백이 아니다 — proposal 시점의 위임 형태에도 같은 공백이 있었다.
    **`not-applicable`: 이 change는 세 분기를 근거로 아무것도 주장하지 않는다.**
  - `replay.go:551`이 **오류를 버린다.** 원장 쓰기가 실패해도 조용하다.
    **a099가 만든 문제가 아니고 a099가 고칠 것도 아니다.**
  - `remindAfter=0`이 재무장을 끄는 것과 임차를 끄는 것을 **한 값에 묶지 않았다.**
    임차를 끄는 것은 값이 아니라 **코드의 부재**다. proposal의 §103 미결 사항은
    §4.7이 이 방향으로 닫았다.
