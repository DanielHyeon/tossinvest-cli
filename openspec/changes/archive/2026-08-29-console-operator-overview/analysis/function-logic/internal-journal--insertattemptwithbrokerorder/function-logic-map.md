# Function Logic Map: `insertAttemptWithBrokerOrder`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트 헬퍼**다(HEAD L257-266). 브로커가 ack한 PLACE 시도 한 건을
`mutation_attempts`에 직접 넣는다 — 쓰기는 **쓰기 handle**(`*Journal`의 `j.db`)로 하고,
read-only handle은 이 파일의 다른 헬퍼가 따로 연다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `j *Journal` | 쓰기 handle(테스트 전용 TempDir) | `openTestJournalAt` | — |
| `attemptID`, `intentID` | 이미 존재하는 intent를 참조 | 호출부(`insertIntent`가 선행) | FK 위반 → `t.Fatalf` |
| `brokerOrderID` | 임의 문자열(빈 문자열 포함) | 호출부 | 빈 값은 "ack 없는 시도"를 만든다 |

고정 컬럼: `kind='PLACE'`, `state='RECORDED'`, `attempt_no=1`, `fingerprint='fp'`,
`recorded_at='2026-03-30T00:30:00Z'` — `schemaV1`의 `NOT NULL` 컬럼을 만족시키는 최소값이며,
테스트가 재는 것은 `broker_order_id` 하나이므로 나머지는 상수로 고정한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L259) | `ExecContext`가 오류 | 없음(트랜잭션 미사용, 단문) | `t.Fatalf("insert attempt %s")` | 자체 실행(`TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | 실패 위치를 호출부로 | — | ast.json calls |
| `j.db.ExecContext` | 시도 1행 삽입 | 오류 즉시 `t.Fatalf` | ast.json calls |
| `context.Background` | 테스트 ctx | — | ast.json calls |

## State mutations and fallbacks

- `t.TempDir()` 안의 테스트 journal에 1행 INSERT. 실계좌·브로커·네트워크 무접촉.
- 프로덕션 코드가 이 헬퍼를 부르지 않는다(`_test.go`).

## Safety conclusion

- Safe edit boundary: 신규 테스트 헬퍼 가산. 기존 함수 무변경.
- High-risk impact: **no** — 테스트 전용이고 실계좌에 닿지 않는다. 다만 이것이 재는 대상은
  High-risk(원장 read)이므로, 헬퍼가 `mutation_attempts`에 **엔진의 정상 쓰기 경로를 우회해**
  직접 INSERT한다는 점은 기록해 둔다: `Attempt` 수명주기를 통하지 않으므로
  `attempt_transitions`가 비고, 이 테스트가 재는 것이 `BrokerOrderIDs`의 SQL뿐이라는 뜻이다.
