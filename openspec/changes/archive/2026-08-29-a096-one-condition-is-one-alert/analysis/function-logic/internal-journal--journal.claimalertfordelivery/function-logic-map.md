# Function Logic Map: `Journal.ClaimAlertForDelivery`

- Source: `internal/journal/outbox.go` (169–229)
- AST evidence: `ast.json` (sha256 `c6612c641a3a…`, 11분기, 반환 10곳)
- Risk scan: `risk-pattern-report.md`

a096이 만든 함수다. 본체는 base `ec29dc72`의 `EnqueueAlert`를 그대로 옮긴 것이고,
더한 것은 **두 번째 반환값 `owed`** 하나다.

a096이 만든 함수다. 본체는 base `ec29dc72`의 `EnqueueAlert`를 그대로 옮긴 것이고,
더한 것은 **`owed` 판정과 창을 넘긴 행의 재무장**이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 공백 아님 | 호출자가 조건에서 조립한다(시계가 아니라) | B1: 오류, 행 없음 |
| `a.Type` | 공백 아님 | `obs.EventType` | B2: 오류, 행 없음 |
| `remindAfter` | ≤0이면 재무장 없음 | 호출자의 정책 | — |
| `alert_outbox.event_key` | UNIQUE | 스키마 제약(`outbox.go:50`) | 같은 key는 두 번째 행을 만들지 않는다 |
| `alert_outbox.state` | PENDING·DELIVERED·ACKNOWLEDGED **또는 그 밖** | CHECK 제약 없음 | `claimOwed` default: owed |
| 트랜잭션 | BeginTx…Commit | `j.db` | B3/B7/B9/B10/B11: 오류, `defer tx.Rollback()` |

불변식 셋:

1. **한 event key는 한 행이다.** base와 동일하다.
2. **판정과 id가 같은 순간의 값이다.** 상태·시각을 id와 같은 SELECT로, 같은 트랜잭션
   안에서 읽는다. 그 사이에 무엇도 끼어들 수 없다.
3. **재무장은 판정과 같은 트랜잭션 안에서 일어난다.** owed를 돌려주고 나서 밖에서 재무장하면
   그 사이에 크래시한 프로세스가 "보내라고 들었지만 행은 종결된" 상태를 남긴다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@173 | `key == ""` | 없음 | `0, false, errors.New(…)` | 기존(위임 경유) |
| B2@176 | `a.Type == ""` | 없음 | `0, false, errors.New(…)` | 기존(위임 경유) |
| B3@183 | `BeginTx` 실패 | 없음 | `0, false, wrapped` | 없음(미진입) |
| B4@194 | `switch` | 없음 | — | 열거만 |
| **B5@195** | 기존 행 조회 성공 | `claimOwed` 호출(순수) | 아래 | a096 신규 6건 |
| **B6@197** | `rearm` | **`UPDATE … SET state='PENDING', last_error=''`** | — | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` |
| B7@202 | 재무장 UPDATE 실패 | 없음(롤백) | `0, false, wrapped` | 없음(미진입) |
| B8@209 | 조회가 `ErrNoRows`가 아닌 오류 | 없음 | `0, false, wrapped` | 없음(미진입) |
| B9@217 | `INSERT` 실패 | 없음(롤백) | `0, false, wrapped` | 없음(미진입) |
| B10@221 | `LastInsertId` 실패 | 없음(롤백) | `0, false, wrapped` | 없음(미진입) |
| B11@224 | `Commit` 실패 | 없음 | `0, false, wrapped` | 기존(정상 커밋으로 진입) |

B5의 반환은 `existing, owed, tx.Commit()`(@208). 삽입 경로의 반환은 `id, true, nil`(@228) —
INSERT가 `AlertPending`을 썼기 때문이고, 상수를 다시 쓰는 것이 아니라 INSERT가 쓴 값을
그대로 반영한다.

**재무장이 이 설계의 핵심이다.** 종결된 행을 PENDING으로 되돌리면 리마인더는 최초 전달과
완전히 같은 경로를 걷는다 — 재시도 예산, gate 잠금, 운영 모드 승격이 모두 그대로다.
새 경로를 만들지 않았기 때문에 a074 계약을 다시 검증할 필요가 없다.

`delivered_at`과 `acknowledged_at`은 재무장 때 **지우지 않는다.** 이전 episode의 기록이고,
운영자가 언제 무엇을 확인했는지가 감사 흔적이기 때문이다. 창의 기준은 둘 중 나중 것이며
그 선택은 `latestStamp`가 한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | key·type 정규화 | 없음 | AST :172,:176 |
| `j.clk.Now` | 창 판정과 `created_at` | 주입 시계 | AST :179 |
| `j.db.BeginTx` | 조회+판정+재무장/삽입을 한 원자 단위로 | B3 | AST :182 |
| `tx.QueryRowContext(SELECT id, state, delivered_at, acknowledged_at …)` | 중복 판정 + 창 판정 | `ErrNoRows`가 정상 경로 | AST :191 |
| `claimOwed` | 순수 판정 | 오류 없음 | AST :196 |
| `tx.ExecContext(UPDATE … PENDING)` | 재무장 | B7 | AST :202 |
| `tx.ExecContext(INSERT …)` | 신규 행 | B9 | AST :213 |
| `tx.Commit` | 세 경로 모두 커밋한다 | B11 | AST :208,:224 |

호출자(CodeGraph): `obs.Notifier.claimAndDeliver`(`notifier.go:226`)와
`Journal.EnqueueAlert`(`outbox.go:120`, `owed`를 버리고 창 0을 넘긴다).

## State mutations and fallbacks

- 신규 행은 항상 `AlertPending`. 전달 표시는 `MarkAlertDelivered`가 한다.
- 창 안의 기존 행은 **아무것도 쓰지 않는다.**
- 창 밖의 종결 행은 `state`와 `last_error`만 쓴다. `attempts`는 누적을 유지한다 —
  그 행이 지금까지 몇 번의 전송 시도를 겪었는지가 사라지면 안 된다.
- 실패 시 `defer tx.Rollback()`이 유일한 되돌림이며 부분 상태가 남지 않는다.

## Safety conclusion

- Safe edit boundary: B5의 SELECT 확장, `claimOwed` 호출, B6의 재무장 UPDATE, 반환값 하나.
- High-risk impact: **yes** — `internal/journal`은 원장이다. 다만 주문·손절·익절·사이징·
  체결 경로에 닿지 않고, 새로 쓰는 열은 `state`와 `last_error` 둘뿐이며 둘 다 알림 전용이다.
- 실질 위험은 `owed`가 **틀리게 false**가 되어 보내야 할 알림을 삼키는 것이다.
  1판의 영구 억제가 정확히 그 형태였고, 독립 리뷰가 두 경로를 제시했다(transport 사망
  미탐지, 다른 사유의 재발 은폐). 창이 둘 다 닫는다.
