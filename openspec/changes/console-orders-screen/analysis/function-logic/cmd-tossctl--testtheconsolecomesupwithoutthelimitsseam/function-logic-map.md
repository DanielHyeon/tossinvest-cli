# Function Logic Map: `TestTheConsoleComesUpWithoutTheLimitsSeam`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=base, L471–485, 분기 2개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — 본문 **byte 동일**. 인접 삽입의 diff hunk 교차로 evidence가 요구되었다 (revision=base)

config 파일을 해석하지 못했다고 콘솔이 **기동을 거부하면** 무엇이 잘못됐는지 알아내려는 바로 그 순간에 콘솔이 없다. seam은 nil이고 개요는 패널마다 그렇게 말한다.

`console.New`에 `StartVerify`만 주고 나머지 seam은 전부 비운 채로 뜨는지 본다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `console.Options{StartVerify: ...}` | 다른 seam 전부 nil | 이 테스트 | New가 에러면 FAIL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `console.New` 에러 | — | `t.Fatalf` | 이 테스트 |
| B2 | `c.Handler() == nil` | — | `t.Error` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `console.New` | seam 없는 콘솔 구성 | nil seam 허용이 계약 | internal/console |
| `c.Handler()` | 핸들러 존재 확인 | — | L526 |

## State mutations and fallbacks

- 테스트 — 서버를 띄우지 않는다.

## Safety conclusion

- Safe edit boundary: 필수 seam 집합. `GateLimits`/`Orders`/`Signals`를 필수로 만드는 변경이 이 테스트를 깬다.
- High-risk impact: no (기동 가능성 검사 — 주문·Guardian·원장 경로 무접촉).
