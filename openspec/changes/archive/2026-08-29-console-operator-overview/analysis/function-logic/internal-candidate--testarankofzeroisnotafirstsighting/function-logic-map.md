# Function Logic Map: `TestARankOfZeroIsNotAFirstSighting`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. 부재는 0이 아니고, 목록 길이가 붙은 rank 0은 '목록 맨 위'가 아니다 — `percentileOf`가 정확히 그렇게 만든다. §1 7절이 `rank > 0 && rank_total == 0`에서 같은 실패를 이미 만났다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 | `Promote("KR","005930", t0)` | 테스트 | 존재하는 후보 |
| 거부되어야 할 순위쌍 | `{0,150} {12,0} {-1,150} {0,0} {151,150}` | 테스트 | 전부 에러 |
| 그 밖의 침묵 불가 사유 | 순간 없음, 원천 없음, 대상 후보 없음 | 테스트 | 전부 에러, 마지막은 `ErrNoCandidate` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 다섯 순위쌍 순회 | — | — | (테스트 자체) |
| B3 | 잘못된 순위쌍이 수용됨 | 없음 | `t.Errorf` | (테스트 자체) |
| B4 | 순간 없는 최초 목격이 수용됨 | 없음 | `t.Error` | (테스트 자체) |
| B5 | 원천 없는 최초 목격이 수용됨 | 없음 | `t.Error` | (테스트 자체) |
| B6 | 모르는 후보에 대한 write가 `ErrNoCandidate`가 아님 | 없음 | `t.Errorf` | (테스트 자체) |
| B7 | 모르는 후보의 `FirstRank`가 found=true거나 기록됨 | 없음 | `t.Errorf` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.NoteFirstRank` | 인자 검증 네 갈래 | 전부 에러 | `store.go:1324` |
| `s.FirstRank` | 두 번째 반환의 의미 | found=false | `store.go:1406` |
| `errors.Is` | `ErrNoCandidate` 확인 | — | ast.json calls |

## State mutations and fallbacks

- 임시 저장소만 만진다. 어떤 단언도 실제로 컬럼을 채우지 않는다 — 전부 거부 경로다.
- `{151,150}`이 목록에 있는 것은 §4 P2가 `PercentileExceeds`가 `Rank > RankTotal`을 막지 않는다고 지적했기 때문이다.

## Safety conclusion

- Safe edit boundary: 다섯 쌍 중 어느 하나라도 빼면 `rank/0 = +Inf`나 100%를 넘는 백분위가 다시 열린다
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
