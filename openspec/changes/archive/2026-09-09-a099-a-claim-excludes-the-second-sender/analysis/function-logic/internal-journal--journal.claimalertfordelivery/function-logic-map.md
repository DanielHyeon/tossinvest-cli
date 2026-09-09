# Function Logic Map: `Journal.ClaimAlertForDelivery`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 산출물이 a099의 전제였다.** proposal 시점 이 함수는 이름만 claim이었고
> PENDING 행에 아무 표시도 남기지 않았다. 그 사실이 design D0의 근거이고,
> **그 주장을 쓰기 전에 이 열거를 만들었다.**
>
> **§5.6 갱신(구현 후).** 그때 분기는 11이었다. 지금은 8이다. **줄어든 셋은 사라진 것이
> 아니라 옮겨 갔다** — §4.7이 기록 절반을 `recordAlertTx`로 뽑았다.
> `internal-journal--recordalerttx` 번들이 그 분기들을 진다.
> 이 함수에 남은 것은 **트랜잭션 껍데기와 두 갈래(기록만 / 기록 + 청구)**다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey`·`a.Type` | 공백 아님 | `alertKey` `:234` | B1 → 이탈 `:236` 오류 (**트랜잭션 전**) |
| **`claimant`** | **공백 아님** | B2 `:238` | 이탈 `:239` 오류. **a099가 더한 인자이자 분기다** |
| `remindAfter` | `<= 0`이면 시간 기반 재무장 없음 | 호출자 (`notifier.go:251` `n.remindAfter()`) | 값이 아니라 정책이다 — 기록만 하는 호출자는 이 함수를 아예 안 부른다(D13) |
| `alert_outbox.state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED`, **CHECK 없음** | `schemaV3` | 미지의 값은 owed·rearm 둘 다 참 (`claimOwed`) |
| **임차 열 넷** | `claim_token`·`claimed_by`·`claimed_at`·`claim_expires_at` | **schemaV31** | **a099가 만든 것이 이 행이다** |
| `j.alertLease` | `> 0` | `Open`이 채운다 (`DefaultAlertLease` 81s) | `:259`가 그대로 넘긴다 |
| 트랜잭션 | `:241` BeginTx · `:245` defer Rollback | — | commit은 `:252`와 `:263` **둘** |

**proposal이 성립하지 않는다고 적은 불변식** — *"claim한 자만 보낸다"* — 을
이 함수가 이제 원장에서 강제한다. 예전 doc comment는 그것을 호출자에게 넘겼다
(*"Exclusion … is the caller's: obs.Notifier holds its delivery mutex"*).
**그 문장은 지워졌다.** 지금 doc comment `:225-230`은 무엇을 사는지와
**무엇을 못 사는지**(전화기가 두 번 울리는 것은 막지 못한다)를 같이 적는다.

## Branches and early returns

AST 열거 — 분기 8 · 이탈 9 · 호출 13 · 대입 6 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:235` | `alertKey` 거절 | **없음 — 트랜잭션 전이다** | `:236` `ClaimResult{}, err` | 기존 |
| **B2 `:238`** | **`claimant`가 공백** | **없음 — 트랜잭션 전이다** | `:239` 오류 | **`TestClaimingWithoutANameIsRefused`** |
| B3 `:242` | `BeginTx` 실패 | 없음 | `:243` 포장 오류 | 없음 |
| B4 `:248` | `recordAlertTx` 실패 | 롤백 | `:249` `err` | 없음 |
| **B5 `:251`** | **`!owed` — 보낼 것이 없다** | commit만 | `:257` **`ClaimSettled`** | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` |
| B6 `:252` | `!owed` 경로의 commit 실패 | 롤백 | `:253` 포장 오류 | 없음 |
| B7 `:260` | `acquireAlertClaimTx` 실패 | 롤백 | `:261` `err` | 없음 |
| B8 `:263` | 청구 경로의 commit 실패 | 롤백 | `:264` 포장 오류 | 없음 |
| — 이탈 `:266` | 정상 | **임차 UPDATE 하나** (`acquireAlertClaimTx` 안) | `out` — `ClaimAcquired` 또는 `ClaimHeldElsewhere` | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` |

### B5가 D13의 두 번째 자리다

`!owed`면 **청구하지 않고** 커밋한다. 보낼 것이 없는 행에 임차를 걸면
아무도 풀지 않을 잠금이 남는다 — `EnqueueAlert`가 이 함수를 안 부르는 것과
같은 이유이고, 여기는 그 이유가 **같은 함수 안에서** 다시 나타나는 자리다.

### 세 분기가 어디로 갔나

| proposal 시점 분기 | 지금 어디 |
|---|---|
| `:194` SELECT switch · `:195` 기존 행 · `:197` rearm | `recordAlertTx` B2·B3·B4 |
| `:229` 재무장 UPDATE 실패 | `recordAlertTx` B5 |
| `:241` `!ErrNoRows` · `:249` INSERT 실패 · `:253` `LastInsertId` 실패 | `recordAlertTx` B6·B7·B8 |
| `:173` key · `:176` type | **합쳐졌다** — `alertKey`가 둘 다 본다 (B1 하나) |

**분기 수가 줄어든 것을 「단순해졌다」로 읽으면 안 된다.** 옮겨 간 곳의 번들이 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `alertKey` `:234` | 입력 검증 + 중복 제거 키 | 트랜잭션 **전** | `ast.json` calls |
| `strings.TrimSpace` `:238` | 청구자 이름 검증 | — | 같음 |
| `j.db.BeginTx` `:241` | 쓰기 트랜잭션 | `_txlock=immediate`. ctx 취소 전파 | 같음 |
| `tx.Rollback` `:245` (defer) | 모든 이른 이탈 | 커밋 뒤에는 no-op | `ast.json` defers |
| `j.recordAlertTx` `:247` | **기록 절반** — dedup·삽입·재무장·`owed` | 오류는 그대로 올린다 | `ast.json` calls |
| `tx.Commit` `:252` | `!owed` 경로 | 실패 시 B6 | 같음 |
| **`acquireAlertClaimTx` `:259`** | **청구 절반 — 판정이 여기 있다** | 오류는 드라이버 오류만. 「못 잡았다」는 오류가 아니라 `ClaimHeldElsewhere` | 같음 |
| `j.clk.Now` `:259` | 임차 발급 시각 | 주입된 시계 | 같음 |
| `tx.Commit` `:263` | 청구 경로 | 실패 시 B8 | 같음 |

**두 절반이 한 트랜잭션 안에 있다.** 기록과 청구 사이에 다른 발송자가 끼어들 창이 없다.

**live binding — 유일한 프로덕션 호출자**: `notifier.go:251` (`claimAndDeliver`).
`EnqueueAlert`는 **더 이상 이 함수를 부르지 않는다** (§4.7, D13).

## State mutations and fallbacks

- **기록 절반의 mutation은 `recordAlertTx` 안이다** — 삽입 또는 재무장.
- **청구 절반의 mutation은 UPDATE 하나** — 임차 열 넷을 쓴다(`alert_claim.go:232-236`).
  판정은 **전부 그 UPDATE의 WHERE 안**이고 Go 쪽에 사본이 없다.
- **`!owed`면 임차를 안 쓴다** (B5). 정산된 행에 잠금을 남기지 않는다.
- **폴백**: 미지의 상태는 owed·rearm 둘 다 참으로 **열린 쪽**으로 실패한다.
  임차도 같은 방향이다 — 만료·역행 시계는 행을 **다시 청구 가능**하게 만든다.
- **재무장은 임차를 지운다** (`recordAlertTx`의 UPDATE에 `alertClaimCleared`).
  새 episode는 새 발송이고, 이전 episode를 잡고 있던 자가 이것을 보내지 않는다.

## Safety conclusion

- **Safe edit boundary**: 판정을 Go로 끌어오는 편집이 이 함수의 금지선이다.
  술어의 사본이 둘이 되면 갈라진다 — `acquireAlertClaimTx`의 doc comment `:210-213`이
  그것을 명시한다. B1·B2를 트랜잭션 **뒤로** 옮기는 편집도 금지다:
  write lock을 잡고 입력을 거절하게 된다.
- **High-risk impact**: **yes** — 원장 스키마이고 진입 게이트가 같은 테이블의
  `UndeliveredCount`에 반응한다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B3·B4·B6·B7·B8에 테스트가 없다.** 전부 드라이버·커밋 실패 경로다.
    **`not-applicable`: 이 change는 다섯을 근거로 아무것도 주장하지 않는다.**
  - **트랜잭션 격리 수준을 이 함수가 고르지 않는다** — `BeginTx(ctx, nil)`은
    드라이버 기본이다. 임차 CAS가 그 기본 위에서 배타적인지는 논증이 아니라
    `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`가 관측한다.
    **`-race` 실행은 §5.2이고 이 문서를 쓰는 시점에 아직 안 돌았다.**
  - **이 함수는 전화기가 두 번 울리는 것을 막지 못한다.** 자기 임차를 넘겨
    멈춰 있던 발송자의 push는 만료 뒤에도 도착할 수 있다. 원장의 보장은
    **누가 정산하는가**에 대한 것이고 몇 번 울리는가에 대한 것이 아니다 —
    doc comment `:225-230`이 그렇게 적는다.
