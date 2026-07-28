# Function Logic Map: `ReadOnly.BrokerOrderIDs`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 메서드**다(HEAD L249-271). 기존 함수의 본문을 고친 것이 아니라 파일 끝쪽에
가산됐고, 같은 파일의 기존 read 넷(`AccountRefs`, `LivePositionExits`,
`AccountExitEvents`, `AccountTradeTrips`)은 무변경이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 요청 수명 | 콘솔 핸들러 | 취소 시 질의 중단 |
| `r.db` | `mode=ro` + `query_only(true)` 단일 커넥션 | `readOnlyDSN` (`readonly.go`) | 쓰기 문장은 `SQLITE_READONLY` |
| `mutation_attempts` 존재 | `OpenReadOnly`가 보장 | `readOnlyTables` (`readonly.go`) | 없으면 handle 자체가 발급되지 않음(`ErrSchemaTooOld`) |
| `broker_order_id` | `TEXT NOT NULL DEFAULT ''` (`schemaV1`) | `internal/journal/schema.go` | 빈 문자열 행은 `WHERE broker_order_id <> ''`로 제외 |

불변식 셋:

1. **쓰기 경로가 없다.** `*ReadOnly`에는 쓰는 메서드가 존재하지 않고, 그 사실은 리뷰 습관이
   아니라 `TestTheReadOnlyHandleHasNoWriteMethods`가 reflect로 **메서드 집합을 열거**해
   지킨다. 이 change는 그 allowlist에 `BrokerOrderIDs` 한 줄을 추가했다.
2. **계좌 스코프가 없다** — 이 파일에서 유일하다. 나머지 넷은 한 계좌의 포지션과 그 이력을
   투영하므로 계좌를 섞으면 틀린 화면이 된다. 이 read는 **브로커가 발급한 주문번호**와
   대조하는 집합이라 스코프가 필요 없고, `intents.account_ref`로 조인하면 철자 차이 하나가
   행을 조용히 떨어뜨려 화면에 "엔진이 낸 주문이 하나도 없다"로 렌더된다.
3. **빈 id는 결과가 아니다.** ack를 못 받은 시도의 기본값이 `''`이고, 빈 항목은 id가 없는
   모든 주문에 매칭되어 전부를 "엔진 발주"로 귀속시킨다.

`readOnlyTables` 확대(같은 change, `readonly.go` L78-80)의 additive 근거는 스키마 이력이다:
`mutation_attempts`는 `schemaV1`, 나머지 넷은 `schemaV6`이고, migration은 forward-only이며
`schema.go`의 규칙 3이 drop/rename을 금지한다. 따라서 넷을 가진 파일은 반드시 그것도 갖는다.
**대가**는 정직하게 적는다: 그 테이블이 없는 journal은 이제 새 read 하나가 아니라
**모든 read**가 open 단계에서 거절된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L254) | `QueryContext` 실패 | 없음 | `journal: listing the broker order ids in <path>` | `TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery`(이 경로에 도달하기 전에 open이 막음) |
| B2 (for, L260) | `rows.Next()` 순회 | 로컬 slice append | — | `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued` |
| B3 (if, L262) | `rows.Scan` 실패 | 없음 | `journal: reading a broker order id` | 동상(컬럼 계약 회귀 시) |
| B4 (if, L267) | `rows.Err()` 비정상 | 없음 | B1과 같은 문장 | 동상 |

정렬·중복 제거·빈 id 제외는 분기가 아니라 SQL(`SELECT DISTINCT … WHERE broker_order_id <> ''
ORDER BY broker_order_id`)이 소유하고, 그 셋을 한 번에 재는 것이
`TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`다(같은 id 두 번 기록 + ack 없는 시도).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.db.QueryContext` | `mutation_attempts` 한 컬럼의 DISTINCT | ctx 전파, retry 없음, `busy_timeout`은 DSN 소유 | ast.json calls |
| `rows.Close` (defer) | 커서 해제 | — | ast.json defers |
| `rows.Scan` / `rows.Err` | 행 읽기·순회 오류 | 첫 오류에서 즉시 반환 | ast.json calls |
| `fmt.Errorf` | 경로를 포함한 문장 | — | ast.json calls |

## State mutations and fallbacks

- DB 변경 0. 커넥션이 `mode=ro`(SQLITE_OPEN_READONLY) + `query_only(true)`라 두 겹으로 거절되고,
  `journal_mode`는 DSN이 건드리지 않는다(설정 자체가 쓰기다).
- 프로세스 상태 변경 0 — 캐시도 두지 않는다.
- fallback 없음. 오류를 빈 slice로 바꾸지 않는다: 이 화면에서 빈 결과는
  "모든 주문이 수동 발주"라는 **주장**이다.

## Safety conclusion

- Safe edit boundary: 신규 메서드 가산 + `readOnlyTables` 1항목 + 메서드 allowlist 1행.
  기존 read 넷과 `OpenReadOnly`/`readOnlyDSN`/`checkSchema` 본문은 무변경.
- High-risk impact: **yes** — 원장 읽기 표면이고, 화면의 "이 주문은 엔진이 냈다" 귀속의
  유일한 근거다. 틀리면 운영자가 엔진이 놀고 있다고 결론 내린다.
  그럼에도 additive인 이유: (a) 쓰기 경로가 타입 수준에서 없고 그 사실을 guard test가 열거로
  지킨다, (b) 커넥션 계약(`mode=ro` + `query_only`)이 그대로다, (c) 기존 네 read의 SQL·시그니처·
  오류 문장이 그대로다, (d) 새로 등록한 테이블이 `schemaV1`이라 기존에 열리던 journal의
  집합이 줄지 않는다.
