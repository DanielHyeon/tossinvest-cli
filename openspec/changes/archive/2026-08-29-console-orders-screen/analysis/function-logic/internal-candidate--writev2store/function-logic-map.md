# Function Logic Map: `writeV2Store`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 헬퍼다. 현장의 저장소 대부분은 v2이므로 v2 → v3가 실제 업그레이드가 밟는 칸이다. fixture의 요점은 두 번째 후보 — 승격 9분 **전**의 148위 행이다. 대칭인 읽기 창 안에 있고, D20이 이름 붙인 두 형태(삶 사이의 간극, 승격 전 표류)가 사는 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v2Schema` | SchemaVersion 2를 그대로 적은 DDL | 같은 파일 상수 | v1Schema와 같은 이유로 유도하지 않는다 |
| 후보 005930 | 기준선 이미 기록됨, 순위 행 2건(t0/12위, t0+1m/8위) | 테스트 | 백필이 t0의 12위를 골라야 한다 |
| 후보 000660 | `first_seen_at = t0+1h`, 순위 행 1건이 그 9분 **전** 148위 | 테스트 | 백필이 **가져가면 안 되는** 행 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `sql.Open` 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B2 | v2 스키마 생성 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B3 | 첫 후보 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B4 | 순위 행 2건 순회 | INSERT | — | (헬퍼) |
| B5 | 순위 행 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B6 | 두 번째 후보 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B7 | 간극 관측 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sql.Open` + `dsn` | 프로덕션과 같은 DSN | — | `store.go:dsn` |
| `db.Exec` | v2 DDL과 seed | 실패는 `t.Fatalf` | (헬퍼) |
| `stamp` | 고정폭 직렬화 | — | 마이그레이션이 실제로 만날 텍스트여야 한다 |

## State mutations and fallbacks

- 임시 파일에 v2 스키마 데이터베이스를 만들고 후보 2건·관측 3건을 넣는다.
- 첫 후보가 이미 기준선을 갖고 있다는 것이 v1 fixture와의 차이다 — v3 rung이 아래 칸이 방금 돈 것에 기대지 않는지 보는 것이 이 fixture의 목적이다.

## Safety conclusion

- Safe edit boundary: 두 번째 후보의 '승격 9분 전' 행을 승격 뒤로 옮기면 D20의 실제 형태가 fixture에서 사라진다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
