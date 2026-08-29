# Function Logic Map: `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`

- Source: `internal/candidate/store_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1663–1803, 분기 22개)
- Risk scan: `risk-pattern-report.md`

v2 저장소가 올라오면서 D20의 최초 순위 백필이 도는 것. 이 change가 **결론을 뒤집었다**.

전에는 백필된 위치가 곧바로 92번째 백분위로 **측정 가능**하다고 단언했다. 지금 단언하는
것은 정직한 쌍이다: 위치는 마이그레이션을 온전히 건넜고 **보이지만**, veto는 이유를 명명하며
그것으로 측정하기를 거부한다.

schema-4 rung이 두 자격 사실을 더하면서 **어느 쪽도 백필하지 않기** 때문이다. 이 행들이
쓰일 때 아무도 그것을 재지 않았다 — 이 저장소가 갖고 있던 boolean은 쓰인 모든 행에서
구조적으로 false였다(어떤 소스도 그 필드를 대입하지 않았으므로) — 그래서 그것을 앞으로
복사하면 "아무도 보지 않았다"가 "소스가 이미 목록에 있다고 했다"로 한 UPDATE에 바뀐다.
그리고 그 모든 위치가 `seen_late`가 측정해도 되는 위치가 된다.

주석은 한 번 더 정정됐다: "업그레이드 한 스캔 뒤 답을 되찾는다"는 것도 **거짓**이다.
회복은 후보의 생명이 끝날 때 일어난다(만료 → `Promote`가 초기화 → 새 읽기가 자격을 준다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v2 fixture | 손으로 적은 v2 스키마 | 이 파일의 `v2Schema` | — |
| 백필 대상 | 동일성 창 안의 가장 이른 순위 행 (12 of 150) | v3 rung | 창 밖 행은 백필되지 않는다 |
| schema-4 rung | **백필 없음** | `migrations[3]` | 기존 행은 미상 — 그것이 일어난 일이다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Open` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 칼럼 목록 순회 | 없음 | — | 자체 실행 |
| B3 | 기대 칼럼이 없다 | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 버전 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | 버전이 현재가 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B6 | 후보 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 | 생애 instant가 바뀌었다 | 없음 | `t.Errorf` | 자체 실행 |
| B8 | provenance가 바뀌었다 | 없음 | `t.Errorf` | 자체 실행 |
| B9 | baseline 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B10 | baseline이 다르다 | 없음 | `t.Errorf` | 자체 실행 |
| B11 | 관측 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B12 | 관측 2행이 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B13 | 최초 순위 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B14 | 백필된 위치가 다르다 | 없음 | `t.Errorf` | 자체 실행 |
| B15 | 마이그레이션 뒤 sighting이 **측정됐다** | 없음 | `t.Errorf` — **이 change가 뒤집은 단언** | 자체 실행 |
| B16 | 사유가 `NEW_ENTRANT_UNKNOWN`이 아니다 | 없음 | `t.Errorf` — 왜 못 재는지 이름이 있어야 한다 | 자체 실행 |
| B17 | 거부된 sighting이 읽은 것을 보고하지 않는다 | 없음 | `t.Errorf` | 자체 실행 |
| B18 | gap 행 후보 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B19 | gap 행이 백필됐다 | 없음 | `t.Errorf` | 자체 실행 |
| B20 | `Close` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B21 | 재오픈 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B22 | 재오픈 후 최초 순위가 다르다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open(...)` | 사다리 실행 | `t.Fatal` | ast.json calls |
| `MeasureFirstSighting(got, first, rows)` | **뒤집힌 단언의 대상** | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 단언 3개 교체(측정 가능 → 명명된 거부 + 읽은 것 보고) + 주석 정정 2회.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — 백필이 미측정을 측정으로 바꾸면 저장된 이력 전체가 `seen_late`에 노출된다.
