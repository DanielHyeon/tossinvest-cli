# Function Logic Map: `TestARankFromOutsideTheIdentityWindowIsNotStored`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 테스트다. `NoteFirstRank` 창 가드의 **쓰기 쪽 절반**을 고정한다. 스캔 루프는 tick마다 순위를 제안하고, 가격이나 캔들로 올라온 후보가 몇 시간 뒤 순위 목록에 잡히는 것은 평범한 일이다. 그것을 저장하면 D17의 늦은 기준선과 같은 실수인데 **안전한 방향이 없다** — 종목은 그동안 올랐을 수도 내렸을 수도 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 | `Promote("KR","005930", t0)` | 테스트 | `first_seen_at = t0` |
| 창 밖 관측 | `t0 + DefaultStalenessTTL`에 4위/150 | 테스트 | 저장되지 않아야 하고 **에러도 아니어야** 한다 |
| 창 안 관측 | `t0 + DefaultStalenessTTL - 1s`에 4위/150 | 테스트 | 저장되어야 한다 — 경계가 말한 자리에 있음을 양쪽에서 고정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 창 밖 write가 **에러를 냄** | 없음 | `t.Fatalf` — 순위를 나르지 않는 원천이 올린 후보는 평범한 상태다 | (테스트 자체) |
| B3 | 반환값이 `Recorded()` | 없음 | `t.Errorf` | (테스트 자체) |
| B4 | 저장소에 남아 있음 | 없음 | `t.Errorf` | (테스트 자체) |
| B5 | 창 안 write 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 창 안 1초가 저장 안 됨 | 없음 | `t.Errorf` — 경계가 말한 자리에 없다 | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.NoteFirstRank` | 창 가드의 쓰기 쪽 | 창 밖은 `FirstRank{}, nil` | `store.go:1376` |
| `s.FirstRank` | 저장소가 실제로 비어 있음을 확인 | — | `store.go:1406` |

## State mutations and fallbacks

- 임시 저장소만 만진다.
- 경계를 **양쪽에서** 건다. 한쪽만 걸면 '아무것도 저장하지 않는' 구현이 통과한다 — §4 P2가 `TestAnAbsentInputAgeLimitIsTheDefaultAndNotNoLimit`에서 만난 실패 형태다.

## Safety conclusion

- Safe edit boundary: 창 안 단언(B5·B6)을 빼면 창 가드를 통째로 지운 구현이 green이 된다 — 금지
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.
