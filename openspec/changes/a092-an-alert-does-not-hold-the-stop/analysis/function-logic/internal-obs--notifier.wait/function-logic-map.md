# Function Logic Map: `Notifier.wait`

- Source: `internal/obs/notifier.go` (410-420)
- AST evidence: `ast.json` — branches 2, returns 1, calls 2, assignments 4,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

34초 예산의 **+4초**가 여기서 나온다. a092가 값을 바꾸는 두 필드 중 하나(`RetryDelay`)의
소비처다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | `deliver`가 받은 것 그대로 | `deliver:388` | 취소되면 `Sleep`이 오류 → `false` |
| `n.RetryDelay` | 0 허용 | `Notifier` 필드(`notifier.go:107`) | B1 `:412`가 0 이하를 `DefaultRetryDelay` **2s**(`:48`)로 대체 |
| `n.Clock` | nil 허용 | `Notifier` 필드(`:102`) | B2 `:416`이 nil을 `clock.System()`으로 대체 |

**프로덕션은 `RetryDelay`를 채우지 않는다** — `newNotifier`(`exitwiring.go:73-81`)의
구조체 리터럴에 그 필드가 없다(`internal-app-engine--newnotifier/ast.json`: branches 0,
returns 1, calls 0). 그래서 실효값은 **2s**다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:412` | `delay <= 0` | `delay = DefaultRetryDelay` (2s) `:413` | — |
| B2 `:416` | `clk == nil` | `clk = clock.System()` `:417` | — |
| — `:419` | — | `clk.Sleep(ctx, delay)` | `return ... == nil` |

**이탈은 하나다.** 잠들지 않고 반환하는 경로가 없다 — ctx가 이미 취소돼 있어도
`clock.Sleep`의 계약에 달렸다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.System` `:417` | 시계 대체 | 즉시 | AST calls |
| `clk.Sleep` `:419` | **대기** | `delay`만큼 잔다. ctx 취소로만 단축된다 | AST calls |

### `deliver`가 이 함수를 부르는 횟수

`deliver` B7 `:387` `if attempt < attempts`이므로 **`attempts - 1`회**.
`attempts` = 3(프로덕션) → **2회** → 2 × 2s = **4s**.

```text
critical 최악 = attempts × (publish 1회 상한) + (attempts-1) × RetryDelay
              = 3 × 10s + 2 × 2s = 34s      ← 오늘
```

## State mutations and fallbacks

- **`Notifier`의 상태를 바꾸지 않는다.** AST assignments 4는 전부 지역 변수(`delay`, `clk`)다.
- fallback은 두 기본값(B1·B2)뿐이다.
- goroutine·defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 이 함수를 **편집하지 않는다.** 바꾸는 것은 호출자가
  채워 주는 `n.RetryDelay` 값이다 — 이 함수의 분기·이탈은 그대로다.
- **High-risk impact**: 간접. 이 대기가 exit 관측 루프의 체류에 그대로 더해진다.
- **`RetryDelay`를 줄일 때의 위험**: 재시도 간격이 짧으면 "잠깐 죽은 transport"가
  회복할 시간을 못 준다. 다만 `Ntfy.Publish`의 실패는 대부분 **timeout**이고
  (`ntfy.go:99` `context.WithTimeout`), timeout은 그 자체로 대기다 —
  간격을 줄여도 시도 사이 실제 경과는 `Timeout + RetryDelay`다.
