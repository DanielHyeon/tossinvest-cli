# Function Logic Map: `Saga.event`

- Source: `internal/flatten/flatten.go` (L689-700)
- AST evidence: `ast.json` (1 branch, 1 return)
- Risk scan: `risk-pattern-report.md`

**a097은 이 함수를 편집하지 않는다.** 산출물이 필요한 이유는 proposal R2가 이 함수를
"알림 오류를 버리는 호출자"의 실증으로 인용하기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Notifier` | nil 가능 | Saga 조립 지점 | B1: 로그만 남기고 반환 |
| `t`, `key`, `fields` | 호출자 | flatten saga의 단계들 | — |
| `context.Background()@694` | 취소 불가 컨텍스트 | 이 함수가 직접 만든다 | saga의 ctx를 쓰지 않는다 |

**불변식**: 이 함수는 **오류를 돌려주지 않는다** (반환값 없음). 따라서 알림 실패를 호출자에게
전할 방법이 구조적으로 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@690 | `s.Notifier == nil` | `s.logf@691` | `return` @692 | 기존 (flatten) |

정상 경로: `s.Notifier.Notify(…)@694`의 결과를 **`_ =`로 버린다.**

**이것이 R2의 근거다.** 알림 실패를 오류 반환만으로 통지하는 설계는 이 호출자 하나로
무효가 된다. 그래서 a097은 호출자가 아니라 **생산자**(`claimAndDeliver`)에서 gate를
잠근다 — 호출자는 앞으로도 늘고, 그때마다 같은 검사를 요구할 수 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.logf@691` | Notifier 미배선 시 로그 | 오류 없음 | AST calls |
| `s.Notifier.Notify@694` | 운영자 알림 | **오류를 버린다** | AST calls |
| `context.Background@694` | saga ctx 대신 | 취소되지 않는다 | AST calls |

## State mutations and fallbacks

- mutation 없음. 로그와 알림 발송뿐이다.
- fallback: Notifier가 없으면 로그로 강등(B1).

## Safety conclusion

- Safe edit boundary: **없음 — a097은 이 파일을 읽기만 한다.**
- High-risk impact: yes (비상 청산 saga의 관측 경로)
- a097 이후 이 함수의 동작은 그대로지만 **결과가 달라진다**: 버려진 오류가 claim 실패였다면
  이제 그 시점에 진입 게이트가 잠겨 있다. 호출자를 고치지 않고 계약을 만족시키는 것이
  design D2의 선택이다.
