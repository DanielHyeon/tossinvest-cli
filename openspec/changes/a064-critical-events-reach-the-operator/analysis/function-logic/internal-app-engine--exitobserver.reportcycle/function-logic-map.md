# Function Logic Map: `ExitObserver.reportCycle`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a064-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: Normal. **새 함수**이며, 편집 대상 파일(`exitloop.go`) 안에 있어 도구가
  수정 함수로 집계한다. `ExitObserver.Run`의 map과 함께 읽는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cycle ExitCycle` | 한 관측 사이클의 결과 | `ObserveOnce` | — |
| `cycle.Err` | nil이면 성공 | `ObserveOnce` | 이 함수의 유일한 분기 조건 |
| `o.opts.Log` | nil 허용 | 배선 | `o.logErr`가 nil 검사를 먼저 한다 |

**불변식 1**: 성공한 사이클은 아무것도 남기지 않는다. 관측 주기는 5초이고, 정상
사이클마다 한 줄을 쓰면 하루 17,280줄이 되어 실패 줄을 덮는다.

**불변식 2**: 알림이 아니라 로그다. `obs.EventExitCycleFailed`는 `criticalEvents`에
없으며, 그것이 이 설계의 핵심이다 — `cycle.Err`는 의미가 하나가 아니고, critical로
만들면 일시적 원장 오류가 outbox → 전달 실패 → entry gate latch → ENTRY_BLOCKED로
이어져 살아 있는 계좌를 멈춘다.

**불변식 3**: 반환값이 없다. 호출자가 이 함수의 결과로 분기할 수 없어야 한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (379) | `cycle.Err == nil` | 없음 | 조기 반환 — 침묵 | 2.2 |

B1의 else 경로(= 실패한 사이클)가 유일한 side effect다: `o.logErr` 한 번.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.logErr` | 실패 사유를 error 등급으로 기록 | nil Logger에서 no-op (`exitloop.go`) | AST |

`o.alert`를 호출하지 **않는다**. 그것이 D1이 코드에서 읽히는 지점이다.

## State mutations and fallbacks

- 상태를 바꾸지 않는다. 맵도 카운터도 만지지 않는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 함수 전체가 a064의 것이다.
- High-risk impact: **no** — 판정·주문·손절 경로에 값을 공급하지 않고 반환값도 없다.
- §0.3: 호출 시점은 사이클이 끝난 뒤이므로 어떤 청산도 지연시키지 않는다.
