# Function Logic Map: `TestPruningRawObservationsLeavesTheFirstRankToo`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. D11의 약속을 task 4.9가 더한 컬럼으로 확장한다. 이 컬럼이 존재하는 바로 그 경우다 — 순위가 온 행이 정확히 보존이 지우는 것이고, `seen_late`는 가장 오래 달린 후보에 관한 질문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 관측 1건 | `t0`, 148/150, 가격 10000 | 테스트 | 보존 창 밖으로 밀어낼 대상 |
| 후보 | `Promote(t0)` + `NoteFirstRank(148,150,t0)` | 테스트 | — |
| prune | `PruneObservations(t0 + DefaultRawRetention)` | 테스트 | 원 행은 전부 사라져야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 관측 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | 최초 순위 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | prune 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 관측 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 원 행이 남아 있음 | 없음 | `t.Fatalf` — 남아 있으면 이 테스트는 아무것도 재지 않는다 | (테스트 자체) |
| B7 | 순위 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B8 | 순위·목록 길이·순간이 유실됨 | 없음 | `t.Errorf` | (테스트 자체) |
| B9 | prune 뒤 `seen_late`가 측정 불가 | 없음 | `t.Errorf` — 컬럼이 존재하는 이유 전부 | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.PruneObservations` | 원 행 삭제 | `observations`만 건드린다 | `store.go:828` |
| `s.FirstRank` | 요약이 살아남았는지 | — | `store.go:1406` |
| `MeasureFirstSighting` | 살아남은 값이 실제로 답을 내는지 | 원 행 slice로 `nil`을 준다 | `veto.go` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- B6이 없으면 prune이 아무것도 안 한 경우에도 green이다 — §4 P2가 지적한 '맞는 이유로 통과하지 않는 테스트' 형태의 예방이다.
- B9가 저장소 단언을 넘어 veto까지 간다. 컬럼이 남은 것과 그 값으로 답이 나오는 것은 다른 주장이다.

## Safety conclusion

- Safe edit boundary: B6이나 B9를 빼면 각각 '아무것도 안 지운 prune'과 '남았지만 쓸 수 없는 값'이 통과한다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
