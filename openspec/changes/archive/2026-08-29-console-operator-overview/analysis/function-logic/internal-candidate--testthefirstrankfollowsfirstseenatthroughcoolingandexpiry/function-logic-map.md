# Function Logic Map: `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. 위 기준선 테스트와 같은 모양을 세 번째 같은 등급의 사실에 적용한다. 재진입에서 떨어뜨리면 D1의 우회가 한 필드 옆으로 열리고 — 5위까지 오른 종목이 한 스캔 목록을 떠났다 돌아와 5위를 '처음 본 자리'로 기록한다 — 만료에서 이어받으면 D20이 재현한 '죽은 삶이 산 삶을 대신 답한다'가 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 005930 | `Promote(t0)` + `NoteFirstRank(148, 150, t0)` | 테스트 | 목록 바닥 근처에서 처음 본 후보 |
| 재진입 | `Promote(t0+2m)` 뒤 5위/150 제안 | 테스트 | 저장소가 이미 있는 값을 지켜야 한다 |
| 만료 | `t0+2m + StalenessTTL + CoolingTTL` | 테스트 | 재승격 없이 staleness만으로 |
| 새 삶 | 5위/150을 `SourceOfficialGainers`가 보고 | 테스트 | `MeasureFirstSighting`이 96.666666666666%를 내야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 최초 순위 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | 냉각 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | 냉각 중 순위가 사라짐 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 재승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 재진입이 `first_seen_at`을 옮김 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 재진입의 5위 제안이 에러가 됨 | 없음 | `t.Fatalf` — 거부는 조용해야 한다 | (테스트 자체) |
| B8 | 재진입 후 순위 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B9 | 재진입이 순위를 5위로 옮김 | 없음 | `t.Errorf` — `seen_late`가 그때부터 clear된다 | (테스트 자체) |
| B10 | 만료 시점 상태 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B11 | 상태가 만료가 아님 | 없음 | `t.Fatalf` | (테스트 자체) |
| B12 | 만료 후 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B13 | 만료 후 순위 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B14 | 새 후보가 죽은 후보의 순위를 상속 | 없음 | `t.Errorf` — D20의 결론 그대로 | (테스트 자체) |
| B15 | 순위의 순간·원천이 만료를 넘어 살아남음 | 없음 | `t.Errorf` | (테스트 자체) |
| B16 | 새 삶의 순위 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B17 | 새 삶의 순위가 5위가 아님 | 없음 | `t.Fatalf` | (테스트 자체) |
| B18 | `seen_late`가 5/150을 반영 못 함 | 없음 | `t.Errorf` — 저장된 값이 실제로 veto까지 간다 | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.Promote` / `s.Cool` / `s.Candidate` | 수명 구동 | — | `store.go` |
| `s.NoteFirstRank` / `s.FirstRank` | 순위 write·read | 이미 있으면 조용히 기존 값 | `store.go` |
| `MeasureFirstSighting` | 저장된 값이 veto까지 도달하는지 | — | `veto.go` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- B18이 이 테스트를 저장소 단언 이상으로 만든다 — 컬럼이 채워진 것과 그 값이 `seen_late`를 답하는 것은 다른 주장이고, 후자가 컬럼이 존재하는 이유다.

## Safety conclusion

- Safe edit boundary: B9(재진입 보존)나 B14(만료 초기화) 중 하나만 남기면 각각 D1의 우회와 D20의 결함이 다시 열린다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
