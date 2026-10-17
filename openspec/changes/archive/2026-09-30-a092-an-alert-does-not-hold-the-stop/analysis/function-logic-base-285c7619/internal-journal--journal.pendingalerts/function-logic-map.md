# Function Logic Map: `Journal.PendingAlerts`

- Source: `internal/journal/outbox.go` (392-405)
- AST evidence: `ast.json` — branches 2, returns 2, calls 5, assignments 3,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용이자 위험 지점.** a092는 이 함수를 편집하지 않지만, **17판이 이 함수를
부르는 방식은 바꾼다.** HEAD `Flush:437`은 `limit = 0`으로 부르고, B1의 조건이
`limit > 0`이므로 **`LIMIT` 절이 붙지 않는다** — 즉 미전달 행 전부다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `limit` | `> 0`이면 `LIMIT`, `<= 0`이면 **무제한** | 호출자 | 없음 — 0이 유효한 입력이다 |
| 대상 행 | `state = PENDING`만 | `:393` `WHERE state = ?` | — |
| 순서 | `ORDER BY id` — 오래된 것 먼저 | `:393` | — |
| `ctx` | 호출자의 것 | 17판에서는 배달 루프의 것 | 취소되면 B2 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| **B1 `:395`** | **`limit > 0`** | 쿼리에 `LIMIT ?` 추가 `:396-397` | — |
| B2 `:400` | `QueryContext` 실패 | 없음 | 오류 `:401` |
| — `:404` | — | 없음 | `scanAlerts(rows)` |

**B1이 17판 D0의 네 속성 중 하나가 필요한 이유다.** `limit = 0`이면 이 함수는
미전달 알림을 **전부** 돌려준다. HEAD `Flush`는 그 전부를 한 잠금 안에서
순회하며 행마다 `Publish`를 부른다(`notifier.go:441-451`). 밀린 행이 9개고
전송 상한이 3.5초면 한 번의 `Flush`가 31.5초 동안 뮤텍스를 쥔다 —
**a092가 없애려는 바로 그 결함이 자리만 옮긴 것이다.**

그래서 17판의 spec은 **사이클당 작업량 상한**을 SHALL로 쓰고,
설계는 `alertFlushBatch = 8`을 이 함수의 `limit`으로 넘기기로 한다.
**이 함수는 바뀌지 않는다. 인자가 바뀐다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append` `:397` | `LIMIT` 인자 | 순수 | AST calls |
| `j.db.QueryContext` `:399` | 조회 | 로컬 SQLite | AST calls |
| `rows.Close` `:403` | **defer** | — | AST defers **1** |
| `scanAlerts` `:404` | 행 → `[]Alert` | 스캔 오류를 올린다 | AST calls |

**네트워크 없음.** 체류는 행 수에 비례하는 로컬 읽기뿐이다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| — | — | **없음. 읽기 전용이다** |

- fallback 없음.
- **`WHERE state = PENDING`이 D0.3의 세 번째 다리다**: 재무장이 건드리는 행은
  `DELIVERED`/`ACKNOWLEDGED`/미상이고, 이 쿼리는 그것들을 애초에 돌려주지 않는다.
  그래서 배달 루프가 손에 쥔 행과 재무장이 덮어쓰는 행은 겹치지 않는다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — **인자 하나가 사이클 체류의 상한을 정한다.**
  17판 구현이 `limit = 0`을 그대로 쓰면 spec의 "사이클당 작업량 상한" SHALL을
  위반한다. §6.0 **R17-5**가 그 상한을 관측한다.
- 읽기 전용이므로 잠금 없이 여러 goroutine이 불러도 안전하다. 그러나
  **읽은 결과가 다음 순간에도 참이라는 보장은 없다** — 재무장이 그 사이에
  새 행을 PENDING으로 만들 수 있다. 17판은 그것을 **다음 주기가 집는다**로
  처리하고, 그 지연을 게이트 래치 시점의 대가로 이미 기록했다.
- **`Acknowledge:491`도 `limit = 0`으로 부른다.** 그쪽은 배달이 아니라
  일괄 승인이므로 전부가 맞다 — 같은 인자가 두 곳에서 다른 의미다.
