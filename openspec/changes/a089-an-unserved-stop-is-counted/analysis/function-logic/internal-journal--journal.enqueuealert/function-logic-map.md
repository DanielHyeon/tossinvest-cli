# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go` (111-151)
- AST evidence: `ast.json` — branches 9, returns 9, calls 20, assignments 7, **defers 1** (`:125`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 비어 있으면 안 됨 | 호출자 | B1에서 오류 |
| `a.Type` | 비어 있으면 안 됨 | 호출자 | B2에서 오류 |
| `alert_outbox.event_key` | `UNIQUE` (`schemaV3`) | 스키마 | 중복 삽입 불가 → 조회 후 재사용 |
| `state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED` | `:71-80` | **B5는 이 값을 읽지 않는다** |

**호출자 2곳**: `obs.Notifier.notifyCritical` (`notifier.go:177`) — 뒤이어 `deliver`를 부른다 ·
`execgw.replay.parkAlert` (`replay.go:551`) — **전달자가 없다**.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | 기존 테스트 |
|---|---|---|---|---|
| B1 `:113` | `key == ""` | 없음 | `return 0, err` `:114` | `TestEnqueueAlertRequiresAKeyAndAType` |
| B2 `:116` | `Type == ""` | 없음 | `return 0, err` `:117` | 같은 테스트 |
| B3 `:122` | `BeginTx` 오류 | 없음 | `return 0, err` `:123` | 없음 |
| B4 `:129` | `switch` (기존 행 조회 결과) | — | — | — |
| B5 `:130` | **기존 행 발견** | **없음 — 상태 무관하게 id만 반환** | `return existing, tx.Commit()` `:131` | `outbox_test.go:45-61` |
| B6 `:132` | `ErrNoRows` 아닌 오류 | 없음 | `return 0, err` `:133` | 없음 |
| B7 `:140` | `INSERT` 오류 | 없음 | `return 0, err` `:141` | 없음 |
| B8 `:144` | `LastInsertId` 오류 | 없음 | `return 0, err` `:145` | 없음 |
| B9 `:147` | `Commit` 오류 | 없음 | `return 0, err` `:148` | 없음 |
| — `:150` | 정상 삽입 | 새 `PENDING` 행 | `return id, nil` | 다수 |

## B5가 장부를 얼린다

기존 행을 찾으면 **그 행의 `state`를 보지 않고** id만 돌려준다. 그 뒤의 모든 갱신은
`state = 'PENDING'`을 요구한다 — `MarkAlertDelivered` `:159`, `MarkAlertAttemptFailed` `:174`,
`AcknowledgeAlert` `:195`. 행이 이미 `DELIVERED`면 셋 다 0행을 만나
`requireOneRow`(`:282-290`)가 `ErrAlertNotFound`를 낸다.

**실측 (2026-08-05, `journal.db` + `engine.log`)**: 거부 5회 → 행 1개(id 12), `attempts=1`,
`no such alert: 12` 로그 **4건**. 첫 건이 행을 만들어 전달했고 나머지 넷이 B5로 들어와
`MarkAlertDelivered`에 막혔다.

**해로운 귀결**: 재발의 전달이 *실패*하면 행이 `DELIVERED`로 남아
`PendingAlerts`(`:210`)에 안 뜨고 `UndeliveredCount`(`:228`)에 안 세어진다 →
`Acknowledge`(`notifier.go:370-376`)의 `remaining == 0`이 자명하게 참이 되어
**운영자가 아무것도 확인하지 않은 채 게이트가 풀린다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.BeginTx` `:121` | 조회+삽입 원자성 | 오류 → B3 | AST |
| `tx.Rollback` `:125` | **defer** | 커밋 후에는 no-op | AST |
| `tx.QueryRowContext` `:128` | `event_key`로 기존 행 조회 | `ErrNoRows` 분기 | AST |
| `tx.ExecContext` `:136` | 삽입 | 오류 → B7 | AST |
| `res.LastInsertId` `:143` | 새 id | 오류 → B8 | AST |
| `tx.Commit` `:131,:147` | 커밋 | 오류 → B9 | AST |

브로커 접촉 없음. 원장 로컬 트랜잭션 1회.

## State mutations and fallbacks

- 정상 경로: `alert_outbox` 1행 삽입 (`state=PENDING`, `attempts=0`)
- **B5 경로: 아무 행도 바뀌지 않는다** — 재발이 장부에 흔적을 남기지 않는다
- fallback 없음. `defer tx.Rollback()`이 미커밋 트랜잭션을 되돌린다

## Safety conclusion

- **Safe edit boundary**: B5에서 재발 시 행을 갱신하는 것은 **원장 쓰기**이고 브로커·판정·
  제출 어디에도 닿지 않는다. 스키마 변경 없이 기존 열로 표현 가능
- **High-risk impact**: **yes** — 원장 경로이며 게이트 해제 술어(`UndeliveredCount`)의 입력
- **폭발 반경**: `parkAlert`(`replay.go:551`)는 **전달자가 없다**. B5를 무조건 재개방으로
  바꾸면 그 경로의 행이 영구 `PENDING`이 되어 게이트가 안 풀린다. 재개방은
  **`notifyCritical` 경로 한정**이어야 한다(옵트인 인자)
- **기존 단언과 충돌**: `outbox_test.go:56-61`이 "최초 관측 본문을 유지한다"를 의도로
  못박아 뒀다. 갱신하려면 그 단언을 뒤집고 근거를 남겨야 한다
