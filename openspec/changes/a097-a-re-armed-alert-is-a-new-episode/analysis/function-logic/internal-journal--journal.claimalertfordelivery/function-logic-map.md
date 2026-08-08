# Function Logic Map: `Journal.ClaimAlertForDelivery`

- Source: `internal/journal/outbox.go` (L169-237)
- AST evidence: `ast.json` (11 branches, 10 returns)
- Risk scan: `risk-pattern-report.md`

a097이 **본문을 바꾸는** 함수다. 바뀌는 것은 재무장 UPDATE(`tx.ExecContext@208`)의
SET 목록이며, 분기 구조와 트랜잭션 경계는 바뀌지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.EventKey` | 공백 아닌 문자열 | 호출자 (`Notifier.eventKey`) | B1에서 error, 트랜잭션 시작 전 |
| `a.Type` | 공백 아닌 문자열 | 호출자 | B2에서 error, 트랜잭션 시작 전 |
| `remindAfter` | `<= 0`이면 시한 재알림 비활성 | `Notifier.remindAfter()` 또는 `EnqueueAlert`의 0 | 판단은 `claimOwed`가 한다 |
| `alert_outbox.state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED` 또는 미지 값 | 원장. **CHECK 제약 없음** | 미지 값은 owed+rearm (`claimOwed` B8) |
| `alert_outbox.event_key` | UNIQUE | 스키마 (`outbox.go` 스키마 정의) | 중복 INSERT 불가 → 조회 후 분기 |
| 시각 | `j.clk.Now()` | 주입된 clock | 미래 스탬프도 `claimOwed` B6이 fail-open |

**불변식**: 상태 조회와 재무장은 **같은 트랜잭션 안**에서 일어난다(`BeginTx@182` …
`Commit@216`). 그래서 "지금 owed인가"와 "그래서 PENDING으로 되돌렸다"가 한 순간의 답이다.
동시 claim에 대한 배타는 이 함수가 아니라 호출자(`obs.Notifier.mu`)의 책임이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@173 | `key == ""` | 없음 | `0, false, error` @174 | 기존 (a096) |
| B2@176 | `a.Type` 공백 | 없음 | `0, false, error` @177 | 기존 (a096) |
| B3@183 | `BeginTx` 실패 | 없음 | `0, false, wrapped` @184 | 없음 — 장애 주입 (`not-applicable`) |
| B4@194 | `switch` (조회 결과 분류) | 없음 | — | 구조 분기 |
| B5@195 | 기존 행 있음 (`err == nil`) | `claimOwed@196` 호출 | B6/커밋으로 이어짐 | 기존 (a096) |
| B6@197 | `rearm == true` | **UPDATE**: `state`, `last_error`, `acknowledged_at`, `acknowledged_by` — **a097이 `title`·`body`·`payload`·`attempts` 추가** | — | **a097 신규** 2.1·2.2·2.3 |
| B7@208 | UPDATE 실패 | 없음 (defer Rollback) | `0, false, wrapped` @213 | 없음 — 장애 주입 (`not-applicable`) |
| B8@217 | 조회가 `ErrNoRows`가 아닌 오류 | 없음 | `0, false, wrapped` @218 | 없음 — 장애 주입 (`not-applicable`) |
| B9@225 | INSERT 실패 | 없음 | `0, false, wrapped` @226 | 없음 — 장애 주입 (`not-applicable`) |
| B10@229 | `LastInsertId` 실패 | 없음 | `0, false, wrapped` @230 | 없음 — 드라이버 장애 (`not-applicable`) |
| B11@232 | 최종 `Commit` 실패 | 없음 | `0, false, wrapped` @233 | 없음 — 장애 주입 (`not-applicable`) |

**early return이 아닌 정상 반환 둘**: `existing, owed, tx.Commit()` @216 (기존 행),
`id, true, nil` @236 (새 행 — INSERT가 `AlertPending`을 썼으므로 항상 owed).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.clk.Now@179` | 재알림 창 계산 기준 시각 | 오류 없음. 주입 clock | AST calls |
| `j.db.BeginTx@182` | 조회+재무장을 한 순간으로 | 실패 시 B3 | AST calls |
| `tx.QueryRowContext@191` / `Scan@191` | 상태·스탬프 조회 | `ErrNoRows`는 신규, 그 외는 B8 | AST calls |
| `claimOwed@196` | **판단 전체** | 순수 함수. 오류 없음 | 별도 FLM |
| `tx.ExecContext@208` | 재무장 UPDATE | 실패 시 B7 | AST calls |
| `tx.ExecContext@221` | 신규 INSERT | 실패 시 B9 | AST calls |
| `tx.Commit@216`, `@232` | 확정 | 실패 시 B11 (신규 경로) | AST calls |
| `tx.Rollback@186` | `defer`. 커밋된 트랜잭션에는 무해 | — | AST calls |

live config binding 없음. 이 함수는 설정을 읽지 않는다.

## State mutations and fallbacks

- **재무장 UPDATE (B6)** — a097의 편집 지점. a096까지의 SET: `state=PENDING`,
  `last_error=''`, `acknowledged_at=NULL`, `acknowledged_by=''`.
  **a097이 더한 것: `title`, `body`, `payload`, `attempts=0`, `last_attempt_at=NULL`,
  `delivered_at=NULL`** (`outbox.go:229-236`).

  이 문서의 초판은 `delivered_at`을 "의도적으로 남긴다"고 적었다. **틀렸고 뒤집혔다** —
  proposal-freeze 리뷰가 반례를 냈다: 본문을 이번 관측으로 덮으면서 이전 전달 시각을
  남기면 그 행은 "지금 담고 있는 내용이 그때 전달됐다"고 말하며 그것은 거짓이다.
  구현은 지우는 쪽이고 이 문단은 구현을 따른다 (design D1).

  남는 것은 행의 **정체성**뿐이다: `event_key`, `event_type`, `severity`, `created_at`.
- **INSERT** — 신규 행은 `AlertPending`으로 들어가고 `created_at`을 기록한다.
- fallback 없음. 실패는 전부 오류로 반환되며 부분 상태를 남기지 않는다(`defer Rollback`).

## Safety conclusion

- Safe edit boundary: **B6의 UPDATE SET 목록만.** 분기 구조·트랜잭션 경계·반환 계약을
  바꾸지 않는다. owed 판단은 `claimOwed`에 있고 a097은 그것을 건드리지 않는다.
- High-risk impact: **yes** — 원장 행이고 알림 전달 경로다. Pre-Edit 선언 대상.
- 안전 방향: 추가되는 SET은 행을 **더 정확하게** 만들 뿐 owed 판단을 바꾸지 않는다.
  owed였던 알림이 unowed가 되는 경로는 생기지 않는다.
