# Function Logic Map: `TestAnAbsentGracePeriodIsTheDefaultAndNotNoGraceAtAll`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. 0이 경계를 끄는 것은 이 패키지가 반복해서 만난 실패다 — `DayHigh` 0이 '고가에 붙음', veto 임계 0이 '임계 없음'(§4 P0), watch 간격 0이 '가능한 한 빨리'. 여기서 미설정 필드는 '만료된 요약을 즉시 전부 삭제'가 되고, 그것은 후보가 냉각되는 순간 D3의 회계를 파괴한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 | `Promote(t0)` + `Cool(t0)` | 테스트 | `t0 + CoolingTTL`에 만료 |
| grace 후보값 | `0`, `-time.Hour` | 테스트 | 둘 다 기본값(48h)으로 떨어져야 한다 |
| 확인 | `at = t0 + CoolingTTL + RawRetention`, grace = 0 | 테스트 | 이번엔 1건이 지워져야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 냉각 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | grace 후보 2종 순회 | — | — | (테스트 자체) |
| B4 | 스윕 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 방금 만료된 요약이 지워짐 | 없음 | `t.Fatalf` — 미설정 경계는 기본값이지 경계 없음이 아니다 | (테스트 자체) |
| B6 | 확인 스윕 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 기본값 적용 뒤에도 안 지워짐 | 없음 | `t.Errorf` — 위 가드는 스윕이 아무것도 안 지워도 통과한다 | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.PruneExpiredCandidates` | grace 기본값 적용 | 0 이하는 `DefaultRawRetention` | `store.go:886` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- B7이 이 테스트를 §4 P2가 지적한 형태에서 구해 낸다 — 거부 단언만 있으면 '스윕이 아무것도 안 지움'이 green이다. 같은 zero grace로 기본값 창을 지난 뒤 실제로 지워지는 것을 함께 본다.

## Safety conclusion

- Safe edit boundary: B7을 빼면 스윕을 통째로 지운 구현이 통과한다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
