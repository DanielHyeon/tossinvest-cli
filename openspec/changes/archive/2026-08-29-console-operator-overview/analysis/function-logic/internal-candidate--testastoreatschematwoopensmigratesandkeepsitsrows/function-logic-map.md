# Function Logic Map: `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. v1 테스트는 두 칸을 오르고, 그것이 바로 **아래 칸이 같은 트랜잭션에서 방금 돈 경우에만 동작하는 rung을 잡지 못하는** 경우다. 현장 저장소 대부분이 v2이므로 v2 → v3가 실제로 밟히는 칸이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v2 파일 | `writeV2Store` | 같은 파일 | 후보 2건, 관측 3건 |
| `Open` | `FixedFSProber(ext4)` | 테스트 | — |
| 기대 | 순위 컬럼 4개, `schema_version = 3`, 기존 행·기준선 무변경, 창 안 12위 백필, 간극 행 무시 | D20 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | v2 파일 Open 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 기대 컬럼 4개 순회 | — | — | (테스트 자체) |
| B3 | 컬럼 누락 | 없음 | `t.Errorf` | (테스트 자체) |
| B4 | schema_version 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 버전이 3이 아님 | 없음 | `t.Errorf` | (테스트 자체) |
| B6 | 후보 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 수명 순간이 옮겨짐 | 없음 | `t.Errorf` | (테스트 자체) |
| B8 | 완전성·강등 플래그가 옮겨짐 | 없음 | `t.Errorf` | (테스트 자체) |
| B9 | 기준선 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B10 | 기존 기준선이 바뀜 | 없음 | `t.Errorf` — 가산·nullable이라 아무것도 다시 쓰지 않는다 | (테스트 자체) |
| B11 | 관측 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B12 | 원 관측 2건이 안 남음 | 없음 | `t.Fatalf` | (테스트 자체) |
| B13 | 최초 순위 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B14 | 백필된 순위가 창 안 12위가 아님 | 없음 | `t.Errorf` | (테스트 자체) |
| B15 | `seen_late`가 92%가 아님 | 없음 | `t.Errorf` — 컬럼 이전 후보도 답할 수 있어야 한다 | (테스트 자체) |
| B16 | 두 번째 후보 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B17 | 승격 9분 전 148위 행이 백필됨 | 없음 | `t.Errorf` — rung이 D17의 것보다 좁은 이유 전부 | (테스트 자체) |
| B18 | Close 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B19 | 재open 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B20 | 재open이 두 번째 등반을 함 | 없음 | `t.Errorf` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open` | v2 → v3 한 칸 | 한 트랜잭션 | `store.go:climb` |
| `candidateColumns` | `pragma_table_info` | — | 같은 파일 |
| `MeasureFirstSighting` | 백필된 순위가 실제로 `seen_late`를 답하게 하는지 | — | `veto.go` |

## State mutations and fallbacks

- 임시 파일 하나를 v2 → v3로 올린다.
- 이 테스트가 v1 테스트와 겹치지 않는 부분은 두 가지다 — 아래 칸 없이 rung 하나만 도는 경우, 그리고 이미 기준선이 있는 행을 rung이 건드리지 않는다는 것.
- B17이 D20의 실제 형태다. 읽기 창은 대칭이어야 한다(스캔이 승격과 관측을 조금 어긋나게 찍을 수 있다). 마이그레이션 시점의 이른 쪽은 삶 사이의 간극과 승격 전 표류가 사는 자리이고, 하류에서 그것을 구별할 방법이 없다.

## Safety conclusion

- Safe edit boundary: B17을 빼면 v3 rung을 D17의 무조건 백필과 같게 만든 구현이 green이 된다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
