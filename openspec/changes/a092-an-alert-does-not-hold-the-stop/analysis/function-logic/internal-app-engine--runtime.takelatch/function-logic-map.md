# Function Logic Map: `Runtime.takeLatch`

- Source: `internal/app/engine/runtime.go` (428-436)
- AST evidence: `ast.json` — branches 1, returns 2, calls 2, assignments 1,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**이 함수가 D2의 감독자 면제를 지탱한다.** 건강 감독 루프는 1초마다 돌지만
(`DefaultHealthInterval = time.Second`, `runtime.go:84`), 이 래치 때문에
**전송을 1초마다 부르지는 않는다.** a092는 이 함수를 편집하지 않는다 — 인용 전용이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `name` | 루프 이름 | `CheckHealth:383`이 `loop.Name`을 넘긴다 | 없음 |
| `r.escalated` | `map[string]bool` | 이 함수(`:434`)와 `clearLatch` | nil map 쓰기는 panic이지만 `NewRuntime`이 만든다 |
| `r.mu` | `sync.Mutex` | 이 함수와 `clearLatch` | 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:431` | `r.escalated[name]` — 이미 승격함 | 없음 | `return false` `:432` | 이미 래치된 루프는 두 번 알리지 않는다 |
| — `:434` | — | `r.escalated[name] = true` | `return true` `:435` | 첫 초과는 알린다 |

**분기는 하나뿐이고 그 하나가 래치다.** AST branches 1이 그것을 열거로 말한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.mu.Lock` `:429` | 래치 맵 보호 | 즉시 | AST calls |
| `r.mu.Unlock` `:430` (defer) | 해제 | 즉시 | AST defers 1 |

**네트워크 호출이 없다.** AST calls 2가 둘 다 뮤텍스다. 이 함수 자체는 예산과 무관하다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `r.escalated[name]` | `:434` | **래치** — 루프 이름당 1회. 해제는 `clearLatch`뿐이고, 그것은 `CheckHealth` B4 `:375`(`consecutive == 0`, 즉 복구)에서만 불린다 |

- fallback 없음. goroutine 없음(AST `go_statements: 0`).

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: no — 이 함수 자체는 알림 경로가 아니다.
- **a092가 여기서 얻는 사실**: 1초 루프가 매초 전송을 부르지 않는다는 것.
  그래서 engine-safety의 예산 기준을 **exit 관측 주기**에 걸고 감독자를 면제해도
  감독자가 매초 4.6초를 쓰는 일은 생기지 않는다. 두절 1회당 1회다.
  **단, 이것이 면제의 유일한 근거다** — `Runtime.escalate` FLM이 보이듯
  `runtime.go:415`의 두 번째 전송 대기에는 기한이 없다.
