# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go` (111-151)
- AST evidence: `ast.json` — branches 9, returns 9, calls 20, assignments 7,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. `analysis/delivery-latency.md`의 측정이
왜 성립하는지가 여기 B4/B5에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 비어 있으면 오류 | 호출자 `notifyCritical:165-176` | B1 `:113` → 오류 `:114` |
| `a.Type` | 비어 있으면 오류 | 같은 위 | B2 `:116` → 오류 `:117` |
| `ctx` | 호출자의 것 | `notifyCritical:177` | `BeginTx`가 ctx를 쓴다 — **취소되면 원장 기록 자체가 실패한다** |
| `j.clk` | 주입 | 프로덕션 `clock.System()` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:113` | `key == ""` | 없음 | 오류 `:114` |
| B2 `:116` | `a.Type` 공백 | 없음 | 오류 `:117` |
| B3 `:122` | `BeginTx` 실패 | 없음 | 오류 `:123` |
| B4 `:129` | `switch` — 기존 행 조회 결과 | — | — |
| **B5 `:130`** | **`err == nil` — 같은 `event_key` 행이 이미 있다** | **INSERT 안 함** | **`return existing, tx.Commit()`** `:131` |
| B6 `:132` | `!errors.Is(err, sql.ErrNoRows)` | 없음 | 오류 `:133` |
| B7 `:140` | INSERT 실패 | 없음 | 오류 `:141` |
| B8 `:144` | `LastInsertId` 실패 | 없음 | 오류 `:145` |
| B9 `:147` | `Commit` 실패 | 없음 | 오류 `:148` |
| — `:150` | — | 새 행 | `return id, nil` |

**B5가 측정을 가능하게 하는 분기다.** 같은 조건이 다시 관측되면 새 행이 생기지 않고
**기존 id가 그대로 돌아온다**. 그 행이 이미 DELIVERED면 호출자의
`MarkAlertDelivered`가 `WHERE ... AND state = PENDING`에서 0행을 갱신해 오류가 되고
(`outbox.go:154-159`), 그 오류가 `deliver:259`의 로그 줄을 만든다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.BeginTx` `:121` | 트랜잭션 | 로컬 SQLite — 밀리초 | AST calls |
| `tx.QueryRowContext` `:128` | 중복 조회 | 로컬 | AST calls |
| `tx.ExecContext` `:136` | INSERT | 로컬 | AST calls |
| `tx.Commit` `:131`·`:147` | 커밋 | 로컬 | AST calls |
| `tx.Rollback` `:125` | **defer** | 커밋 후에는 무해 | AST defers **1** |

**네트워크 없음.** 이 함수의 체류는 34초 예산에 포함되지 않는다 —
`notifyCritical`의 34초는 전부 `:182` `deliver` 쪽이다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` 새 행 | `:136` | 내구 기록. **critical 알림의 durability는 여기서 끝난다** |
| 기존 행 | B5 `:130` (반환 `:131`) | **변경 없음** — `attempts`도 `created_at`도 안 건드린다 |

- fallback 없음. 실패는 오류로 올라가고 `notifyCritical` B3 `:178`이 호출자에게 돌려준다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 알림의 내구성이 여기 있다. **a092가 예산을 줄여도
  이 함수는 항상 먼저, 그리고 전부 실행된다**(`notifyCritical` `:177`이 `:182`보다 앞).
  즉 **예산 축소가 내구성을 줄이지 않는다** — 줄어드는 것은 발송 시도의 시간뿐이다.
- B5의 부작용(재발이 `attempts`에 안 남는다)은 **a089의 대상**이고 a092의 범위가 아니다.
