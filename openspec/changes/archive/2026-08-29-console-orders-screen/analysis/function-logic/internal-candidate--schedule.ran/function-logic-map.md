# Function Logic Map: `Schedule.Ran`

- Source: `internal/candidate/source.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 이 change의 수정 대상이 아니며, 바로 뒤에 삽입된 `UntilNextDue`의 diff hunk가 이 함수 범위와 교차해 evidence가 요구되었다. base와 현재의 본문은 byte 동일하고 (base L268-272 = 현재 L268-272), ast.json은 그 사실을 고정하려고 base revision으로 붙였다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` SourceID | 패널의 원천 id | 호출자 | 정규화하지 않는다 — 이 map은 `Schedule` 안에서만 쓰인다 |
| `at` time.Time | 읽은 순간 | 호출자 인자(주입 clock) | `at.UTC()`로 저장한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | `s.lastRun[id] = at.UTC()` | 없음(void) | `TestASlowSourceDoesNotHoldBackAFastOne` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.mu.Lock` / `s.mu.Unlock` | `lastRun` map 보호 | — | `YieldToEngine`이 구조적으로 다른 goroutine이라 락이 없으면 Go의 복구 불가 `concurrent map read and map write`다(§2-4) |
| `at.UTC` | 저장 표준화 | — | ast.json calls |

## State mutations and fallbacks

- `Schedule.lastRun` 인메모리 map 하나. 영속되지 않는다.
- 본문 무변경이므로 이 change가 만든 상태 변화는 없다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — 인메모리 스케줄 기록. 주문 경로 무접촉이고 본문은 byte 동일하다.
