# Function Logic Map: `TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 바로 뒤에 삽입된 task 4.9 절(순위 쪽 쌍둥이 테스트 네 개)의 diff hunk가 이 함수와 교차해 evidence가 요구되었다. base L980-1051 = 현재 L980-1051이고 본문은 byte 동일하다. ast.json은 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 005930 | `Promote(t0)` + `NoteFirstPrice("10000", t0)` | 테스트 | 기준선이 있는 활성 후보 |
| 냉각 | `Cool(t0+1m)` | 테스트 | 기준선은 살아 있어야 한다 |
| 재진입 | `Promote(t0+2m)` | 테스트 | `first_seen_at`도 기준선도 그대로여야 한다 |
| 만료 | `t0+2m + StalenessTTL + CoolingTTL` | 테스트 | 아무도 재승격하지 않았으므로 staleness 시계만으로 만료에 닿는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 기준선 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B3 | 냉각 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | 냉각 중 기준선이 사라짐 | 없음 | `t.Fatalf` | (테스트 자체) |
| B5 | 재승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 재진입이 `first_seen_at`을 옮김 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 재진입 후 기준선 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B8 | 재진입이 기준선을 새로 만듦 | 없음 | `t.Errorf` — 이미 두 배 오른 종목이 `extended`를 영영 통과한다 | (테스트 자체) |
| B9 | 만료 시점 상태 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B10 | 상태가 만료가 아님 | 없음 | `t.Fatalf` | (테스트 자체) |
| B11 | 만료 후 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B12 | 만료 후 `first_seen_at`이 새 pass가 아님 | 없음 | `t.Fatalf` | (테스트 자체) |
| B13 | 만료 후 기준선 조회 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B14 | 새 후보가 죽은 후보의 기준선을 상속 | 없음 | `t.Errorf` — 두 삶을 가로질러 재는 것 | (테스트 자체) |
| B15 | 기준선의 순간·원천이 만료를 넘어 살아남음 | 없음 | `t.Errorf` | (테스트 자체) |
| B16 | 새 삶의 기준선 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B17 | 새 삶이 자기 기준선을 못 가짐 | 없음 | `t.Errorf` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.Promote` / `s.Cool` / `s.Candidate` | 수명 구동 | — | `store.go` |
| `s.NoteFirstPrice` / `s.Baseline` | 기준선 write·read | — | `store.go` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- 본문 무변경이므로 이 change가 만든 동작 변화는 없다. 이 change가 한 일은 이 테스트 **옆에** 순위 쪽 쌍둥이(`TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`)를 같은 모양으로 추가한 것이다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
