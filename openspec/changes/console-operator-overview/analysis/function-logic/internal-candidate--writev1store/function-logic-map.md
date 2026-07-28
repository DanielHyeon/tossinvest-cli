# Function Logic Map: `writeV1Store`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 헬퍼다. 이 change가 더한 것은 (1) 같은 후보에 순위 행 2건 — 정체성 창 안(t0, 12위)과 30분 뒤(4위) — 과 (2) 유일하게 살아남은 순위 행이 1시간 늦은 두 번째 후보다. v3 백필이 **고를 것과 남길 것**을 둘 다 갖게 하려는 fixture다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 임시 파일 경로 | 호출 테스트 | `t.TempDir` 아래 |
| `v1Schema` | SchemaVersion 1을 **그대로 적어 둔** DDL | 같은 파일의 상수 | 현재 소스에서 유도하면 v2 파일을 마이그레이션해 놓고 v1이라 부르게 된다 |
| 후보 005930 | `first_seen_at = t0`, 가격 행 2건, 순위 행 2건(t0/12위, t0+30m/4위) | 테스트 | 백필이 t0의 12위를 골라야 한다 |
| 후보 000660 | `first_seen_at = t0`, 순위 행 1건이 t0+1h/9위 | 테스트 | 창 밖이라 백필이 **남겨야** 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `sql.Open` 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B2 | v1 스키마 생성 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B3 | 첫 후보 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B4 | 가격 행 2건 순회 | INSERT | — | (헬퍼) |
| B5 | 가격 행 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B6 | 순위 행 2건 순회 | INSERT | — | (헬퍼) |
| B7 | 순위 행 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B8 | 두 번째 후보 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |
| B9 | 두 번째 후보의 늦은 순위 행 seed 실패 | 없음 | `t.Fatalf` | (헬퍼) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sql.Open` + `dsn` | 프로덕션과 같은 DSN(escape·`_txlock=immediate`)으로 연다 | — | `store.go:dsn` |
| `db.Exec` | v1 DDL과 seed | 실패는 `t.Fatalf` | (헬퍼) |
| `stamp` | 프로덕션과 같은 고정폭 직렬화 | — | 다른 폭으로 쓰면 마이그레이션이 실제로 만날 텍스트가 아니다 |

## State mutations and fallbacks

- 임시 파일에 v1 스키마의 SQLite 데이터베이스를 만들고 후보 2건·관측 5건을 넣는다. 프로덕션 코드는 만지지 않는다.
- 부분 실패는 `t.Fatalf`라 테스트가 즉시 죽는다 — 반쯤 만들어진 fixture로 마이그레이션을 재는 일이 없다.

## Safety conclusion

- Safe edit boundary: `v1Schema`를 현재 소스에서 유도하도록 바꾸면 마이그레이션 테스트가 실패할 수 없게 된다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
