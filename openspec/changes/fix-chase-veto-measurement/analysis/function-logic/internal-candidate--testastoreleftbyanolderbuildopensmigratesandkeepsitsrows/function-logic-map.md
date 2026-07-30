# Function Logic Map: `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1417–1549, 분기 25개)
- Risk scan: `risk-pattern-report.md`

v1 저장소가 사다리 셋을 한 트랜잭션에 올라 현재 스키마가 되고 행을 잃지 않는다는 것.
이 change가 이 테스트에 더한 것은 **`NoteFirstRank` 호출 1곳의 인자 형태**와, 그 위 사다리에
schema-4 rung이 하나 생겼다는 사실이다.

**칼럼 목록 단언(B2·B3)이 schema-4의 네 칼럼을 명명하지 않는다** — 그것은 이 change가
`schema_four_test.go`를 새로 만든 이유다. v1이 세 rung을 한 번에 오르는 것은 "아래 rung이
방금 옆에서 돌았을 때만 작동하는 rung"을 잡지 못하고, 현장의 대부분은 3 → 4 한 걸음이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v1 fixture | 손으로 적은 v1 스키마 | 이 파일의 `v1Schema` | 현재 소스에서 파생하지 않는다 — 그러면 현재 파일을 옛것이라 부르는 것 |
| 행 | 관측 4행 + 후보 2건 | fixture | 전부 살아남아야 한다 |
| `storedFirstRank(...)` | 자격 채움 | `veto_test.go` | B22의 write-once 확인용 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Open` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 칼럼 목록 순회 | 없음 | — | 자체 실행 |
| B3 | 기대 칼럼이 없다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 스키마 버전 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | 버전이 현재가 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B6 | 후보 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 | `first_seen_at`이 바뀌었다 | 없음 | `t.Errorf` | 자체 실행 |
| B8 | sources가 보존되지 않았다 | 없음 | `t.Errorf` | 자체 실행 |
| B9 | provenance 수가 다르다 | 없음 | `t.Errorf` | 자체 실행 |
| B10 | 관측 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B11 | 관측 4행이 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B12 | baseline 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B13 | baseline이 백필되지 않았다 | 없음 | `t.Errorf` | 자체 실행 |
| B14 | 나중 가격 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B15 | 나중 가격이 baseline을 덮었다 | 없음 | `t.Errorf` | 자체 실행 |
| B16 | 최초 순위 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B17 | 최초 순위 백필이 틀렸다 | 없음 | `t.Errorf` | 자체 실행 |
| B18 | 두 번째 후보 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B19 | 한 시간 늦은 행이 백필됐다 | 없음 | `t.Errorf` — 순위는 틀릴 안전한 방향이 없다 | 자체 실행 |
| B20 | 백필이 없는 후보가 측정 가능하다 | 없음 | `t.Errorf` | 자체 실행 |
| B21 | 나중 순위 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B22 | 나중 순위가 백필을 덮었다 | 없음 | `t.Errorf` — write-once | 자체 실행 |
| B23 | `Close` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B24 | 재오픈 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B25 | 재오픈 후 baseline이 다르다 | 없음 | `t.Errorf` — 재오픈은 두 번째 등반이 아니다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open(ctx, Options{...})` | 사다리 실행 | `t.Fatal` | ast.json calls |
| `s.db.QueryContext(PRAGMA table_info)` | 칼럼 목록 | — | ast.json calls |
| `s.NoteFirstRank(...)` | write-once 확인 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 1곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — 마이그레이션이 행을 잃으면 `first_seen_at`이 사라진다.
