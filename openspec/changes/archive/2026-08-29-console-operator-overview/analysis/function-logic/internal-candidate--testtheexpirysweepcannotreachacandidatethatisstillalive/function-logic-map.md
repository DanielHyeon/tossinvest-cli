# Function Logic Map: `TestTheExpirySweepCannotReachACandidateThatIsStillAlive`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. 스윕이 건드리면 안 되는 삶 셋을 세우고, 셋째가 눈에 안 띄는 쪽이다 — 아무도 재승격하지 않은 후보는 `last_seen_at + staleness`에서 **암묵적으로** 냉각하므로 만료 순간이 어느 컬럼에도 없다. `cooled_at`만 보고 쓴 스윕은 그 후보에 영영 닿지 못하거나, 뻔한 수리('cooled_at IS NULL이면 냉각 안 됨')를 하면 살아 있는 후보를 지운다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 3건 | active·cooling·implicit, 전부 `Promote(t0)` | 테스트 | — |
| active | `Promote(t0 + 2×RawRetention)` | 테스트 | 계속 보이는 후보 |
| cooling | `Cool(t0 + 2×RawRetention)` | 테스트 | 방금 냉각된 후보 |
| implicit | 아무것도 안 함 | 테스트 | `t0+staleness`에 암묵 냉각하고 `+cooling`에 만료 — 유예 창을 훨씬 지났다 |
| 스윕 | `at = t0 + 2×RawRetention + 1m`, grace = RawRetention | 테스트 | 정확히 1건이어야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 세 심볼 승격 순회 | — | — | (테스트 자체) |
| B2 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | active 재승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | cooling 냉각 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 스윕 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 삭제 수가 1이 아님 | 없음 | `t.Errorf` | (테스트 자체) |
| B7 | 살아 있어야 할 둘 순회 | — | — | (테스트 자체) |
| B8 | 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B9 | 살아 있어야 할 후보가 사라짐 | 없음 | `t.Errorf` — 산 후보의 `first_seen_at`이 이 패키지의 주장 전부다 | (테스트 자체) |
| B10 | 암묵 냉각한 후보가 살아남음 | 없음 | `t.Errorf` — `cooled_at`만 보는 스윕이 못 닿는 자리 | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.PruneExpiredCandidates` | 두 분기(`cooled_at IS NOT NULL` / `IS NULL`)를 한 번에 검증 | 단일 DELETE | `store.go:897` |
| `s.Candidate` | 생존·소멸 확인 | — | `store.go:1635` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- 세 삶을 한 스윕으로 재는 것이 요점이다 — 하나만 세우면 '아무것도 안 지움'과 '전부 지움' 중 하나가 통과한다.

## Safety conclusion

- Safe edit boundary: B9와 B10을 함께 유지해야 스윕의 두 분기가 각각 필요하다는 것이 고정된다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
