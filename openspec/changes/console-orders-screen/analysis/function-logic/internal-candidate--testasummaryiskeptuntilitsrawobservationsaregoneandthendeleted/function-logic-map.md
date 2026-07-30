# Function Logic Map: `TestASummaryIsKeptUntilItsRawObservationsAreGoneAndThenDeleted`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. D11은 보존을 둘로 나누고 두 번째 절반에 집행자를 주지 않았다 — 기본 동작이 원장의 파일시스템 위에서 무한 증가였다. 유예 기간이 임의값이 아니라는 것이 이 테스트의 주장이다: 원 관측이 사라지면 요약은 아무것과도 조인되지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 | `Promote(t0)` + `Cool(t0+1m)` | 테스트 | 삶이 `t0+1m+CoolingTTL`에 끝난다 |
| 경계 −1ns | `PruneExpiredCandidates(expiry + RawRetention - 1ns, RawRetention)` | 테스트 | 0건이어야 한다 |
| 경계 | `PruneExpiredCandidates(expiry + RawRetention, RawRetention)` | 테스트 | 1건이어야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 냉각 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | 이른 스윕 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | 1ns 이르게 지워짐 | 없음 | `t.Errorf` — 설명할 원 행이 아직 디스크에 있다 | (테스트 자체) |
| B5 | 요약이 이미 사라짐 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 경계 스윕 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 경계에서 1건이 안 지워짐 | 없음 | `t.Errorf` — 아무것과도 조인되지 않는 요약이 영원히 자란다 | (테스트 자체) |
| B8 | 지워졌어야 할 요약이 남음 | 없음 | `t.Errorf` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.PruneExpiredCandidates` | 요약 스윕 | 단일 DELETE | `store.go:883` |
| `s.Candidate` | 실제로 사라졌는지 | found 플래그로 확인 | `store.go:1635` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- 경계를 **양쪽에서** 건다. 이른 쪽만 걸면 아무것도 안 지우는 구현이 통과하고, 늦은 쪽만 걸면 즉시 지우는 구현이 통과한다.
- 이르게 지우는 방향이 D11이 split을 쓴 이유 그대로다 — `first_seen_at`이 그것을 설명하는 이틀치 행이 아직 있는 동안 사라진다.

## Safety conclusion

- Safe edit boundary: 두 경계 단언 중 하나를 빼면 반대편 극단의 구현이 green이 된다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
