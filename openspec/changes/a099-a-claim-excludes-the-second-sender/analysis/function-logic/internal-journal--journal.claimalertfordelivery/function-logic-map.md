# Function Logic Map: `Journal.ClaimAlertForDelivery`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 산출물이 a099의 전제다.** 이름은 claim이라고 말하는데 PENDING 행에 대해
> 아무 표시도 남기지 않는다. 그 사실이 proposal·design D0의 근거이고,
> **그 주장을 쓰기 전에 이 열거를 만들었다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 공백 아님 | 호출자 | B1 `:173` → 이탈 `:174` 오류 |
| `a.Type` | 공백 아님 | 호출자 | B2 `:176` → 이탈 `:177` 오류 |
| `remindAfter` | `<= 0`이면 재무장 없음 | 호출자 (`notifier.go:244` `n.remindAfter()`) | 값이 아니라 정책이다 — 기록만 하는 호출자는 0을 준다 |
| `alert_outbox.state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED`, **CHECK 없음** | `schemaV3` `outbox.go:41-68` | 미지의 값은 owed·rearm 둘 다 참 (`claimOwed` B8 `:308`) |
| **임차 열** | **없다** | 같은 스키마 | **a099가 더하는 것이 이 행이다** |
| 트랜잭션 | `:182` BeginTx · `:186` defer Rollback | — | commit은 이탈 `:240`과 `:256` |

**불변식 하나가 오늘 성립하지 않는다**: *"claim한 자만 보낸다."*
doc comment `:166-168`이 그것을 호출자에게 넘긴다 —
*"Exclusion against a concurrent claimer is the caller's: obs.Notifier holds its
delivery mutex across the claim and the send."*

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 11 · 이탈 10 · 호출 23 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:173` | event key 공백 | 없음 | `:174` 오류 | 기존 |
| B2 `:176` | type 공백 | 없음 | `:177` 오류 | 기존 |
| B3 `:183` | BeginTx 실패 | 없음 | `:184` 오류 | 기존 |
| B4 `:194` | SELECT 결과 분기 (switch) | — | — | — |
| B5 `:195` | `err == nil` — **기존 행이 있다** | `claimOwed@:196` 판정 | (아래로) | **a099 R1** |
| **B6 `:197`** | **`rearm`** | **이 분기 안이 이 경로의 유일한 UPDATE다** | — | **a099 R1 — 오늘 PENDING이면 여기가 거짓이다** |
| B7 `:229` | 재무장 UPDATE 실패 | 없음(rollback) | `:237` 오류 | 기존 (a097) |
| — 이탈 `:240` | B6이 거짓이면 **여기로 바로 온다** | **없음 — 행이 그대로다** | `existing, owed, tx.Commit()` | **a099 R1** |
| B8 `:241` | `!errors.Is(err, sql.ErrNoRows)` | 없음 | `:242` 오류 | 기존 |
| B9 `:249` | INSERT 실패 | 없음 | `:250` 오류 | 기존 |
| B10 `:253` | `LastInsertId` 실패 | 없음 | `:254` 오류 | 기존 |
| B11 `:256` | commit 실패 | 없음 | `:257` 오류 | 기존 |
| — 이탈 `:260` | 새 행 | `:245` INSERT — `:248`이 `AlertPending`을 쓴다 | `id, true, nil` | 기존 |

**B6이 이 change의 자리다.** `claimOwed`가 PENDING에 대해 `rearm=false`를 주므로
(그 번들 B2 `:276` → 이탈 `:278`), 발송을 앞둔 가장 흔한 경우에 **행은 손대지 않은 채**
`owed=true`만 나간다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.clk.Now` `:179` | 트랜잭션의 시각 기준 | 주입된 시계 — 테스트가 가짜를 쓴다 | `ast.json` calls |
| `j.db.BeginTx` `:182` | 트랜잭션 | ctx 취소 전파 | 같음 |
| `tx.QueryRowContext` `:191` + `Scan` `:191` | id·state·두 타임스탬프를 **한 순간에** 읽는다 | `sql.ErrNoRows`가 INSERT 경로를 고른다 | 같음 |
| **`claimOwed` `:196`** | **owed·rearm 판정 전체** | 순수 함수 — 이 change의 판정이 여기 있다 | 같음 |
| `tx.ExecContext` `:229` | 재무장 UPDATE | 실패면 B7 → rollback | 같음 |
| `tx.ExecContext` `:245` | 새 행 INSERT | `event_key` UNIQUE가 중복을 막는다 | 같음 |
| `tx.Commit` `:240`, `:256` | 두 이탈 | — | 같음 |

**live binding — 유일한 프로덕션 호출자**: `notifier.go:244` (`claimAndDeliver`).
`EnqueueAlert`(`outbox.go:120`)가 `remindAfter=0`으로 위임하고 `owed`를 버린다;
그 경로의 호출자는 `replay.go:551` 하나다.

## State mutations and fallbacks

- **PENDING 기존 행: mutation 없음.** 이것이 반증의 내용이다.
- 재무장(B6 참): state·title·body·payload·attempts·last_error·last_attempt_at·
  delivered_at·acknowledged_at·acknowledged_by를 **전부** 되돌린다 (`:229-236`).
  **a099는 여기에 임차 열 초기화를 더해야 한다** (task 4.6) — 새 episode가 이전
  episode의 임차를 물려받으면 안 된다.
- 새 행: `:248`이 `AlertPending`을 쓴다. 발송된 것은 없다 (`:259`의 주석).
- **폴백**: 미지의 상태는 owed·rearm 둘 다 참으로 **열린 쪽**으로 실패한다.
  a099의 임차도 같은 방향이어야 한다 (design D4).

## Safety conclusion

- **Safe edit boundary**: a099는 **B6 `:197`이 거짓인 경로에 UPDATE 하나를 더한다.**
  B1·B2·B3·B8·B9·B10·B11의 조건과 반환은 안 바꾼다. 재무장 UPDATE(`:229-236`)에
  열 넷을 더한다(design C1). 편집 후 AST의 branches가 12 이상이고 **B1~B3의 줄 의미가 그대로면**
  입력 검증 경로 무변화다.
- **High-risk impact**: **yes** — 원장 스키마이고 진입 게이트가 이 함수 아래의
  `UndeliveredCount`에 반응한다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **트랜잭션 격리 수준을 이 함수가 고르지 않는다** — `BeginTx(ctx, nil)`이므로
    드라이버 기본이다. a099의 CAS가 그 기본 위에서 성립하는지는 **R1·R3이
    관측해야 하고 논증으로 대신하지 않는다.**
  - `event_key` UNIQUE 충돌과 임차 CAS 실패는 **다른 사건**이다. 후자는 오류도
    `owed=false`도 아니라 **`ClaimHeldElsewhere`**다 (design C3 · task 4.3).
    **2·3판이 여기서 갈렸다** — 2판 tasks는 `owed=false`라고 적었고 그것이
    2라운드 A-P1이 깬 자리다.
