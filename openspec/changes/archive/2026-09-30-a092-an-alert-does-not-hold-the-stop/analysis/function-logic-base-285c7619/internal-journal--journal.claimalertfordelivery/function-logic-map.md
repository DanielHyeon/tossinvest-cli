# Function Logic Map: `Journal.ClaimAlertForDelivery`

- Source: `internal/journal/outbox.go` (169-261)
- AST evidence: `ast.json` — branches 11, returns 10, calls 23, assignments 10,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판 설계 D0.3이 *"기록 경로가 배달
잠금을 잡지 않아도 이중 발송이 생기지 않는다"*고 주장하고, 그 주장의 절반이 여기
B5·B6의 열거에 있다. 나머지 절반은 `MarkAlertDelivered`/`MarkAlertAttemptFailed`의
CAS 술어에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 공백이면 오류 | 호출자 `claimAndDeliver:244` | B1 `:173` → 오류 `:174` |
| `a.Type` | 공백이면 오류 | 같은 위 | B2 `:176` → 오류 `:177` |
| `remindAfter` | `<= 0`이면 재알림 없음 | `Notifier.remindAfter():280-285` (기본 1시간) | `claimOwed` B4로 전달 |
| `ctx` | 호출자의 것 | `claimAndDeliver`의 ctx | `BeginTx`가 쓴다 — 취소되면 B3 |
| 기존 행의 `state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED`/미상 | `alert_outbox` | `claimOwed`가 판정 |

**트랜잭션 안에서 상태를 읽는다**(`:191` `tx.QueryRowContext`). 그래서 id 해석과
상태 판정은 한 순간의 답이고 그 사이에 아무것도 끼어들 수 없다. 이것은 함수의 doc
comment `:165-168`이 이미 적은 것이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:173` | `key == ""` | 없음 | 오류 `:174` |
| B2 `:176` | `a.Type` 공백 | 없음 | 오류 `:177` |
| B3 `:183` | `BeginTx` 실패 | 없음 | 오류 `:184` |
| B4 `:194` | `switch` — 기존 행 조회 결과 | — | — |
| **B5 `:195`** | **`err == nil` — 같은 `event_key` 행이 있다** | `claimOwed` 호출 `:196` | `:240` `existing, owed, tx.Commit()` |
| **B6 `:197`** | **`rearm == true`** | **UPDATE `:229-236` — state·title·body·payload·attempts·last_error·last_attempt_at·delivered_at·acknowledged_at·acknowledged_by 전부 덮어씀** | — |
| B7 `:229` | 그 UPDATE 실패 | 없음(롤백) | 오류 `:237` |
| B8 `:241` | 조회가 `ErrNoRows`가 아닌 오류 | 없음 | 오류 `:242` |
| B9 `:249` | INSERT 실패 | 없음 | 오류 `:250` |
| B10 `:253` | `LastInsertId` 실패 | 없음 | 오류 `:254` |
| B11 `:256` | `Commit` 실패 | 없음 | 오류 `:257` |
| — `:260` | — | 새 `PENDING` 행 | `id, true, nil` |

**B6이 D0.3의 핵심이다.** 내용을 덮어쓰는 유일한 경로가 B6이고, B6은 `rearm`이
참일 때만 실행된다. `claimOwed`의 `case AlertPending: return true, false`
(`outbox.go:276-278`)가 **PENDING 행에 대해 `rearm=false`를 돌려준다.** 따라서
**PENDING 행은 이 함수가 내용을 덮어쓰지 않는다** — `owed=true`만 돌려주고
`:240`에서 커밋한다.

그리고 배달 루프가 집는 대상은 `PendingAlerts`가 `WHERE state = PENDING`으로
고른 행뿐이다(`outbox.go:393`). 재무장은 `DELIVERED`/`ACKNOWLEDGED`/미상에서만
일어나고, 그 상태의 행은 `PendingAlerts`가 애초에 돌려주지 않는다.

**이것은 주장의 형태이지 관측이 아니다.** 두 경로가 실제 시간축에서 어떻게
끼어드는지는 R17-3이 관측한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:172`·`:176` | 입력 정규화 | — | AST calls |
| `j.clk.Now` `:179` | 재알림 창 계산의 기준 시각 | 주입 시계 | AST calls |
| `j.db.BeginTx` `:182` | 트랜잭션 | 로컬 SQLite — 밀리초 | AST calls |
| `tx.QueryRowContext` `:191` | 기존 행 조회 | 로컬 | AST calls |
| **`claimOwed` `:196`** | **owed/rearm 판정 전부** | 순수 함수 — 오류 없음 | 별도 산출물 `internal-journal--claimowed/` |
| `tx.ExecContext` `:229` | 재무장 UPDATE | 로컬 | AST calls |
| `tx.ExecContext` `:245` | INSERT | 로컬 | AST calls |
| `res.LastInsertId` `:252` | 새 id | 로컬 | AST calls |
| `tx.Commit` `:240`·`:256` | 커밋 | 로컬 | AST calls |
| `tx.Rollback` `:186` | **defer** | 커밋 후 무해 | AST defers **1** |

**네트워크 없음. goroutine 없음**(AST `go_statements: null`). 이 함수의 체류는
전부 로컬 SQLite다 — 17판이 관측 사이클에 남기기로 한 "outbox 트랜잭션 하나"가
바로 이 호출이다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` 새 행 | `:245` | 내구 기록. state는 `PENDING` |
| 기존 행 전 컬럼 | B6 `:229-236` | **재무장** — a097이 "새 에피소드"로 정의한 덮어쓰기 |
| 기존 행 (재무장 아님) | B5, `rearm=false` | **변경 없음** — 읽고 커밋만 |

- fallback 없음. 모든 실패는 오류로 올라가고 `defer tx.Rollback()`이 부분 쓰기를
  남기지 않는다.
- **B6이 `delivered_at`을 `NULL`로 되돌린다**(`:233`). a097의 결론이고, 17판의
  배달 루프가 그 행을 다시 집는 근거다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — critical 알림의 중복 발송 방지가 여기 있다.
  17판이 `n.mu`를 기록 경로에서 떼려면 **이 함수의 B5/B6 열거와
  `MarkAlert*`의 CAS가 함께** 그 배제를 대신해야 한다. 둘 중 하나라도 아니면
  17판의 D0.3은 성립하지 않는다.
- **미해결로 남기는 것**: 두 claimer가 같은 `event_key`로 동시에 들어오는 경우.
  SQLite의 트랜잭션 격리가 그것을 직렬화한다는 것은 **여기서 증명하지 않았다**.
  R17-3이 관측 대상으로 지고 간다.
- B1·B2·B3·B7·B8·B9·B10·B11은 프로덕션에서 도달하려면 SQLite가 실패해야 한다.
  a092는 그 실패 주입을 만들지 않는다(`not-applicable`: 이 change는 이 함수를
  편집하지 않는다).
