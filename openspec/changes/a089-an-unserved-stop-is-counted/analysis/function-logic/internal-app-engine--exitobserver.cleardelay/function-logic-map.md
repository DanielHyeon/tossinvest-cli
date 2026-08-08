# Function Logic Map: `ExitObserver.clearDelay`

- Source: `internal/app/engine/exitloop.go` (1595-1598)
- AST evidence: `ast.json` — branches 0, returns 0, calls 2, assignments 0
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | 비어 있어도 안전 | 호출자 | 없는 키 삭제는 no-op |

**호출자는 저장소 전체에서 하나뿐이다**: `record:1150`.

## Branches and early returns

분기 0. 무조건 두 map에서 키를 지운다.

| Branch | Condition | Mutation/side effect | Return/error | 기존 테스트 |
|---|---|---|---|---|
| — | 없음 | `delete(delayedSince)`, `delete(delayAlerted)` | 없음 | `:914-919` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `delete` `:1596` | `delayedSince` 해제 | 내장, 오류 없음 | AST |
| `delete` `:1597` | `delayAlerted` 해제 | 내장, 오류 없음 | AST |

## State mutations and fallbacks

두 map에서 키 삭제. 그 외 없음.

## Safety conclusion

- **Safe edit boundary**: 함수 자체는 자명하다. **위험은 유일한 호출 지점(`record:1150`)의
  의미**에 있다 — 그것은 "심볼이 깨끗하다"이지 "보호 주문이 살아났다"가 아니다
- **High-risk impact**: **yes (호출 지점 기준)** — 이 한 줄이 브로커 거부 경로의
  지연 알림을 구조적으로 발화 불가로 만든다(a089 1라운드 C1)
- **단독 제거 금지**: `record` B7이 시작한 시계의 유일한 해제점이다. 옮기려면
  `submit`의 `StateConfirmed`가 대신 받아야 하고 `:914-919`가 그 회귀 테스트다
