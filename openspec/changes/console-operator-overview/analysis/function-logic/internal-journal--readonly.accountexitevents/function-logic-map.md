# Function Logic Map: `ReadOnly.AccountExitEvents`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json` (revision `base` — 이 함수는 base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — **본문 무변경**이다. 이 change는 이 함수 **바로 아래**(base L224 뒤,
HEAD L225)에 `ReadOnly.BrokerOrderIDs`를 삽입했고, `git diff --unified=0`의
`@@ -224,0 +225,48 @@` hunk가 함수 끝줄(L223)과 인접해 evidence가 요구됐다.
base(`137cc8d`)의 L188-223과 HEAD의 L188-223은 **바이트 동일**하다
(함수 구간 sha256 `e55e3d72921d8112…` 양쪽 일치, 본 세션에서 추출·비교).
doc 주석도 무변경이다.

이 함수에 대한 이 change의 **유일한 실질 영향**은 함수 밖에 있다: `readonly.go`의
`readOnlyTables`에 `mutation_attempts`가 추가되면서, 그 테이블이 없는 journal은
이제 `OpenReadOnly` 단계에서 `ErrSchemaTooOld`로 거절된다 — 즉 **이 read 를 포함한
모든 read 메서드**가 열리지 않는다. 새 메서드 하나만 실패하는 것이 아니다.
그 확대가 실제로 무해한 근거는 스키마 이력이다: `mutation_attempts`는 `schemaV1`이고
목록의 나머지 넷(`positions`, `exit_states`, `exit_events`, `trade_outcomes`)은
`schemaV6`이며, migration은 forward-only이고 규칙 3이 drop/rename을 금지한다
(`internal/journal/schema.go`) — v6 이상인 파일은 v1 테이블을 반드시 갖는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | 임의 문자열(공백 trim) | 콘솔이 넘긴 masked account ref | 매칭 0행 → 빈 slice |
| `limit` | `>0`만 질의; `<=0`은 즉시 `nil, nil` | 호출자 | 조회 자체를 하지 않음 |
| `r.db` | `mode=ro` + `query_only(true)` 커넥션 | `readOnlyDSN` (`readonly.go`) | 쓰기 문장은 드라이버가 거절 |
| 스키마 | `OpenReadOnly`가 `readOnlyTables` 전부 존재를 확인한 뒤에만 handle 발급 | `checkSchema` | `ErrSchemaTooOld`/`ErrSchemaTooNew` |

불변식: 창(window)은 **최신 끝**을 유지한다(내부 `ORDER BY created_at DESC, id DESC LIMIT ?`
뒤에 바깥에서 오름차순 재정렬). 반대로 자르면 오늘 아침의 ratchet이 화면 밖으로 밀린다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, base L190) | `limit <= 0` | 없음 — 질의도 하지 않음 | `nil, nil` | `TestTheExitHistoryWindowKeepsTheNewestEnd` 계열(account_views_test.go) |
| B2 (if, base L204) | `QueryContext` 실패 | 없음 | `journal: listing the exit history of …` | 스키마 거절은 `OpenReadOnly` 쪽에서 선차단 |
| B3 (for, base L210) | `rows.Next()` 순회 | 로컬 slice append만 | — | account_views_test.go 행 단언 |
| B4 (if, base L212) | `rows.Scan` 실패 | 없음 | `journal: reading an exit event of …` | 컬럼 계약 회귀 시 |
| B5 (if, base L219) | `rows.Err()` 비정상 | 없음 | B2와 같은 문장 | 동상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.db.QueryContext` | exit_events ⨝ positions 투영 | ctx 취소 전파, retry 없음, `busy_timeout`은 DSN이 소유 | ast.json calls |
| `rows.Scan` / `rows.Err` / `rows.Close`(defer) | 행 읽기와 해제 | 첫 오류에서 즉시 반환 | ast.json defers |
| `strings.TrimSpace` | account ref 정규화 | 순수 | ast.json calls |

## State mutations and fallbacks

- 프로세스 상태 변경 없음. DB 변경 없음 — 커넥션이 `mode=ro`이고 `query_only(true)`라
  쓰기는 pager에 닿기 전에 거절된다.
- fallback 없음. 오류는 그대로 올라가며 화면이 "0행"으로 바꿔 읽지 않는다.

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 인접 삽입(`BrokerOrderIDs`)과, 함수 밖
  `readOnlyTables` 한 줄 확대만.
- High-risk impact: **yes** — 원장(ledger) 읽기 경로다. 본문은 무변경이지만 이 change가
  `readOnlyTables`를 넓혔으므로 **이 함수가 호출 가능해지는 전제 조건**이 바뀌었다.
  안전한 이유: 넓어진 조건은 `schemaV1` 테이블 하나이고, 목록의 나머지는 `schemaV6`이며
  migration은 forward-only·drop 금지다. 즉 이 함수가 지금까지 열렸던 모든 journal은
  앞으로도 열린다. 거절이 늘어나는 유일한 경우는 손으로 만들었거나 부분 복원된 파일이고,
  그것은 화면이 조용히 0행을 보여주면 안 되는 바로 그 경우다.
