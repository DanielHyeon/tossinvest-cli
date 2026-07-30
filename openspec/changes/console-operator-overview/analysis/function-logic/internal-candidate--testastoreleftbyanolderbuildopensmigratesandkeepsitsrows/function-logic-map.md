# Function Logic Map: `TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 테스트다. 이 change가 더한 것은 v3 rung의 단언들 — 네 순위 컬럼의 존재, 창 안 행(12위)이 백필되고 30분 뒤 행은 안 된다는 것, 창 밖 순위밖에 없는 후보가 아무것도 얻지 못한다는 것, 그리고 백필된 값도 그 뒤로는 1회성이라는 것. 이 테스트는 사다리를 **두 칸** 오르는 유일한 경우다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v1 파일 | `writeV1Store` | 같은 파일 | 후보 2건, 관측 5건 |
| `Open` | `FixedFSProber(ext4)` | 테스트 | 마이그레이션은 Open 안에서 한 트랜잭션으로 |
| 기대 | 컬럼 7개 추가, `schema_version = 3`, 기존 행 무손실, 두 백필의 서로 다른 규칙 | 설계 D17·D20 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | v1 파일 Open 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 기대 컬럼 7개 순회 | — | — | (테스트 자체) |
| B3 | 컬럼 누락 | 없음 | `t.Errorf` | (테스트 자체) |
| B4 | schema_version 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 버전이 3이 아님 | 없음 | `t.Errorf` — 두 칸 사다리가 순서대로 올랐는지 | (테스트 자체) |
| B6 | 후보 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | `first_seen_at`이 옮겨짐 | 없음 | `t.Errorf` — 이 패키지가 지키는 유일한 필드 | (테스트 자체) |
| B8 | 출처가 유실됨 | 없음 | `t.Errorf` | (테스트 자체) |
| B9 | 완전성 카운터가 유실됨 | 없음 | `t.Errorf` | (테스트 자체) |
| B10 | 관측 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B11 | 원 관측 4건이 안 남음 | 없음 | `t.Fatalf` — 가산·nullable이라 아무것도 다시 쓰지 않는다 | (테스트 자체) |
| B12 | 기준선 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B13 | 백필된 기준선이 가장 오래된 가격 행이 아님 | 없음 | `t.Errorf` | (테스트 자체) |
| B14 | 백필 뒤 write 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B15 | 나중 읽기가 백필된 기준선을 덮음 | 없음 | `t.Errorf` | (테스트 자체) |
| B16 | 최초 순위 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B17 | 백필된 순위가 창 안 12위가 아님 | 없음 | `t.Errorf` — v3 rung은 D17의 rung보다 좁다 | (테스트 자체) |
| B18 | 두 번째 후보 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B19 | 1시간 늦은 행이 최초 목격으로 백필됨 | 없음 | `t.Errorf` — 순위는 틀릴 안전한 방향이 없다 | (테스트 자체) |
| B20 | 그 후보의 `seen_late`가 측정 가능해짐 | 없음 | `t.Errorf` — 미측정이 정직한 답이다 | (테스트 자체) |
| B21 | 백필 뒤 순위 write 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B22 | 나중 읽기가 백필된 순위를 덮음 | 없음 | `t.Errorf` | (테스트 자체) |
| B23 | Close 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B24 | 재open 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B25 | 재open이 두 번째 등반을 함 | 없음 | `t.Errorf` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open` | 마이그레이션 사다리 실행 | 한 트랜잭션. 실패하면 반쯤 오른 파일이 남지 않는다 | `store.go:climb` |
| `candidateColumns` | `pragma_table_info` | — | 같은 파일 |
| `s.Baseline` / `s.FirstRank` | 두 백필의 결과 | — | `store.go` |
| `MeasureFirstSighting` | 백필이 남긴 후보가 정말 미측정인지 | — | `veto.go` |

## State mutations and fallbacks

- 임시 파일 하나를 v1 → v3로 올린다. 프로덕션 상태는 없다.
- 가산·nullable이 WORKFLOW §0.6의 선호이고, 이 테스트가 그것을 '아무 행도 다시 쓰이지 않았다'로 고정한다 — 사다리가 `first_seen_at`을 잃을 수 없다는 뜻이다.
- 두 백필의 규칙이 다르다는 것이 이 테스트의 핵심 단언이다. 가격 백필은 살아남은 가장 오래된 행을 그냥 가져가고(늦음은 `first_price_at`으로 읽기 때문에 잡힌다), 순위 백필은 정체성 창을 **쓰기 시점에도** 적용한다.

## Safety conclusion

- Safe edit boundary: B19·B20(창 밖 후보가 아무것도 얻지 않음)을 빼면 v3 rung을 v2 rung과 같게 만든 구현이 통과한다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
