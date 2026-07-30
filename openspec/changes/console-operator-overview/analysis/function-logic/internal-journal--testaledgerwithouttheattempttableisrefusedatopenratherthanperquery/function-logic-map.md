# Function Logic Map: `TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L305-332). `readOnlyTables`에 `mutation_attempts`를 등록한 이유를
직접 잰다: 등록하지 않으면 `OpenReadOnly`는 **성공하고** `BrokerOrderIDs`만 한 문장씩
실패하며, 실패한 질의를 화면이 0행으로 렌더하면 그것은 "이 계좌의 모든 주문은 손으로
넣은 것"이라는 문장이 된다. 거절은 한 번, open에서, 이름을 대며 일어나야 한다.

**이 change가 그 목록을 넓힌 결과**도 여기서 함께 확정된다: 이제 그 테이블이 없는 journal은
새 read 하나가 아니라 **read 다섯 전부**가 open 단계에서 `ErrSchemaTooOld`로 거절된다.
그 확대가 실무적으로 무해한 근거는 스키마 이력이다 — `mutation_attempts`는 `schemaV1`,
목록의 나머지 넷은 `schemaV6`, migration은 forward-only이고 `schema.go` 규칙 3이
drop/rename을 금지한다. 테스트가 그 테이블을 없애기 위해 `PRAGMA foreign_keys = OFF`와
`DROP TABLE` 두 번을 써야 했다는 사실 자체가 그 근거의 실측이다: 엔진의 FK가 그 테이블을
가리키므로 **정상 journal은 그것을 잃을 수 없다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal 파일 | `t.TempDir()/DBFileName`, 최신 스키마로 생성 | `openTestJournalAt` | `t.Fatalf` |
| 파괴 절차 | `PRAGMA foreign_keys = OFF` → `DROP TABLE attempt_transitions` → `DROP TABLE mutation_attempts` | 이 테스트 | 각 문장 실패 시 `t.Fatalf` |
| 기대 | `OpenReadOnly`가 `ErrSchemaTooOld`, 메시지가 `mutation_attempts`를 이름으로 부른다 | `checkSchema` (`readonly.go`) | `t.Fatalf`/`t.Errorf` |

불변식: 거절은 **타입이 있는 오류**여야 하고(`errors.Is`), 메시지가 무엇이 없는지 말해야 한다 —
"엔진을 한 번 돌려라"는 이유를 볼 수 있을 때만 실행 가능한 지시다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range, L312) | 파괴 문장 3개 순회 | 테스트 DB의 테이블 2개 삭제 | — | 자체 실행 |
| B2 (if, L317) | 파괴 문장이 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 (if, L321) | `j.Close()` 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 (if, L326) | `OpenReadOnly`의 오류가 `ErrSchemaTooOld`가 아님 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 (if, L329) | 메시지가 `mutation_attempts`를 이름으로 부르지 않음 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournalAt` | 최신 스키마 journal 생성 | 실패 시 `t.Fatalf` | ast.json calls |
| `j.db.ExecContext` | FK 해제 + DROP ×2 | 동상 | ast.json calls |
| `OpenReadOnly` | 측정 대상 — open 단계 거절 | `ErrSchemaTooOld` | ast.json calls |
| `errors.Is` / `strings.Contains` | 타입과 문장 확인 | — | ast.json calls |

## State mutations and fallbacks

- `t.TempDir()` 안의 테스트 DB에서만 DROP한다. 실계좌·운영 journal 무접촉.
- `PRAGMA foreign_keys = OFF`는 테스트 픽스처 구성 전용이며 프로덕션 경로에 없다.

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **yes** — 이 테스트가 재는 대상은 원장 handle의 **개방 조건**이고,
  이 change가 실제로 넓힌 것이 바로 그 조건이다. 안전한 방향인 이유: 넓힌 결과는
  "조용한 0행" 대신 "이름을 댄 거절"이며, 넓힌 항목이 `schemaV1` 테이블이라 지금까지 열리던
  journal 집합은 줄지 않는다. 거절이 새로 생기는 경우는 손으로 만들었거나 부분 복원된
  파일뿐이고, 그것이야말로 화면이 조용히 0행을 보여주면 안 되는 경우다.
