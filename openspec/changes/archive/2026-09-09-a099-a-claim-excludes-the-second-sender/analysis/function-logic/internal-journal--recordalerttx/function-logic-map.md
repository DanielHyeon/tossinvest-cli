# Function Logic Map: `Journal.recordAlertTx`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **`check_analysis.py`가 이 번들을 요구하지 않는다.** frozen base에 이 이름이 없으므로
> 「수정된 기존 함수」가 아니다. 그런데도 만든 이유는 **분기가 여기로 옮겨 왔기 때문**이다.
> `ClaimAlertForDelivery`의 11개 분기 중 여섯이 이 함수 안에 있고, a097·a096의
> 재무장 주장은 전부 그 여섯을 근거로 쓴다. **근거가 옮겨 갔는데 증거가 안 따라가면
> 그 근거는 다시 「기억」이 된다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 호출자가 이미 연 쓰기 트랜잭션 | 두 호출자 (`EnqueueAlert` `:146`, `ClaimAlertForDelivery` `:247`) | 롤백은 호출자의 defer가 진다 |
| `a Alert` | `EventKey`·`Type` 공백 아님 | `alertKey` `:279` | B1 → 이탈 `:281` |
| `remindAfter` | `<= 0`이면 **시간 기반** 재무장 없음 | 호출자 | `EnqueueAlert`는 항상 `0` |
| `alert_outbox.event_key` | UNIQUE | schemaV3 | SELECT가 0행 또는 1행 |
| `j.clk.Now()` | 주입된 시계 | `:283` | — |

**불변식**: *"이 함수는 임차를 잡지 않는다."* doc comment `:269-275`가 그렇게 적고,
AST의 calls 14개 어디에도 `acquireAlertClaimTx`가 없다. **그것이 D13의 구조적 근거다.**

## Branches and early returns

AST 열거 — 분기 8 · 이탈 7 · 호출 14 · 대입 8 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:280` | `alertKey` 거절 | 없음 | `:281` `0, false, err` | 기존 (다른 진입점) |
| B2 `:292` | SELECT 결과로 갈리는 `switch` | — | — | — |
| B3 `:293` | `err == nil` — **기존 행이 있다** | `claimOwed` `:294`가 owed·rearm 판정 | (아래) | `TestEnqueueAlertIsIdempotentOnTheEventKey` |
| **B4 `:295`** | **`rearm`** | **이 경로의 유일한 UPDATE** (`:334-342`) | — | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` |
| B5 `:334` | 재무장 UPDATE 실패 | 없음 (호출자가 롤백) | `:343` 포장 오류 | 없음 |
| — 이탈 `:346` | B4가 거짓이면 **바로 여기** | **없음 — 행이 그대로다** | `existing, owed, nil` | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` |
| B6 `:347` | `!errors.Is(err, sql.ErrNoRows)` | 없음 | `:348` 포장 오류 | 없음 |
| B7 `:355` | INSERT 실패 | 없음 | `:356` 포장 오류 | 없음 |
| B8 `:359` | `LastInsertId` 실패 | 없음 | `:360` 포장 오류 | 없음 |
| — 이탈 `:363` | 새 행 | `:351` INSERT — `AlertPending`을 쓴다 | `id, true, nil` | `TestClaimingAFreshAlertOwesDelivery` |

### B4의 UPDATE가 a099의 자리다

재무장은 **열 열둘**을 되돌린다: `state`·`title`·`body`·`payload`·`attempts`·
`last_error`·`last_attempt_at`·`delivered_at`·`acknowledged_at`·`acknowledged_by`
그리고 **`alertClaimCleared`가 펼치는 임차 열 넷**(`:340`).

임차를 안 지우면 **새 episode가 이전 episode의 임차를 물려받는다.**
그 임차를 쥔 발송자는 이미 죽었을 수 있고, 그러면 새 episode는
만료(81초)까지 아무도 못 집는다. 주석 `:328-333`이 그 이유를 적는다 —
*"no rule anywhere says the lease is the exception to 'every column points at one event'"*.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `alertKey` `:279` | 검증 + dedup 키 | 두 호출자가 **이미 한 번 불렀다** — 여기는 키를 다시 얻기 위한 재호출이다 | `ast.json` calls |
| `j.clk.Now` `:283` · `RFC3339` `:284` | `created_at` | 주입된 시계 | 같음 |
| `tx.QueryRowContext` + `Scan` `:289` | id·state·두 타임스탬프를 **한 순간에** | `sql.ErrNoRows`가 INSERT 경로를 고른다 | 같음 |
| **`claimOwed` `:294`** | **owed·rearm 판정 전체** | 순수 함수 | 같음 |
| `tx.ExecContext` `:334` | 재무장 UPDATE | 실패면 B5 | 같음 |
| `errors.Is` `:347` | `ErrNoRows` 판별 | — | 같음 |
| `tx.ExecContext` `:351` | 새 행 INSERT | `event_key` UNIQUE가 중복을 막는다 | 같음 |
| `res.LastInsertId` `:358` | 새 id | 실패면 B8 | 같음 |

**live bindings — 호출자 둘, 둘 다 같은 파일**: `EnqueueAlert` `:146` (`remindAfter=0`,
`owed`를 버린다) · `ClaimAlertForDelivery` `:247` (`owed`를 보고 청구 여부를 정한다).

## State mutations and fallbacks

- **세 갈래**: 재무장 UPDATE(B4) · 아무것도 안 함(B4 거짓) · INSERT(SELECT 0행).
- **PENDING 기존 행은 손대지 않는다.** `claimOwed`가 PENDING에 `rearm=false`를 준다.
  발송을 앞둔 가장 흔한 경우에 이 함수는 **행을 안 건드리고** `owed=true`만 돌려준다 —
  **그래서 a099 이전에는 두 발송자가 같은 답을 받았다.** 배제는 여기가 아니라
  `acquireAlertClaimTx`에 있다.
- **`remindAfter=0`이 상태 복구까지 끄지는 않는다.** 인식 못 하는 state는
  `claimOwed`가 owed·rearm 둘 다 참으로 준다(**열린 쪽 실패**). `EnqueueAlert`의
  주석 `:142-145`가 그 구분을 적는다.
- **폴백 없다.** 실패는 전부 오류로 나가고 호출자의 defer가 롤백한다.

## Safety conclusion

- **Safe edit boundary**: 이 함수에 `acquireAlertClaimTx` 호출을 더하는 편집이 금지선이다.
  그러면 `EnqueueAlert`가 임차를 잡게 되고 D13이 깨진다 —
  **분리의 이유가 인자가 아니라 호출 부재라는 것이 doc comment `:272-275`다.**
  B4의 UPDATE에서 `alertClaimCleared`를 빼는 편집도 금지다.
- **High-risk impact**: **yes** — 원장 쓰기이고, 재무장이 critical 알림의 재발송을 만든다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B5·B6·B7·B8에 테스트가 없다.** 드라이버 오류 경로다.
    **`not-applicable`: 이 change는 넷을 근거로 아무것도 주장하지 않는다.**
  - **B1은 이 진입점으로 관측되지 않는다.** 두 호출자가 이미 `alertKey`를 통과시킨
    뒤에만 여기 도달하므로, **실제로는 도달 불가능에 가깝다.** 그런데도 남아 있는
    이유는 이 함수가 `tx`만 받고 키를 안 받기 때문이다 — 인자를 하나 줄인 대가다.
  - `claimOwed`의 내부 분기는 이 표에 없다. `internal-journal--claimowed` 번들이 진다.
