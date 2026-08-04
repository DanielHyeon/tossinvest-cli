# Function Logic Map: `ExitObserver.Run`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a064-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: **High-risk** — exit 관측 루프의 수명 자체다. 이 함수가 반환하면 손절
  평가가 멈춘다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소되면 루프 종료 | 호출자(runtime supervisor) | `ctx.Err()` 반환 — 정상 종료 |
| `o.ObserveOnce(ctx)` | 한 사이클의 결과 `ExitCycle` | 자기 자신 | 실패는 `cycle.Err`에 담기고 **반환되지 않는다** |
| `o.Interval()` | > 0 (기본 5초) | `opts.Interval` | 0 이하는 기본값으로 대체 |

**불변식 1 (유지)**: 사이클 실패로 반환하지 않는다. 함수 주석이 명시한다 — "a failed
cycle is the hold exit-policy specifies (관측 실패 시 … 판정은 보류된다), and a loop
that exited on one would remove the protection it exists to provide."

**불변식 2 (현재 깨져 있다)**: `ExitCycle.Err`의 선언 주석은 *"It is reported and not
returned by Run"* 이라고 쓰여 있다. **보고하는 코드가 없다.** 현재 본문은
`_ = o.ObserveOnce(ctx)`이고, exit 루프는 runtime에 `Health` 없이 등록되어 있어
(`cmd/tossctl/engine.go:373-380`) 감독자도 사이클 실패를 셀 수 없다.

**a064가 바꾸는 것**: 반환된 `ExitCycle`을 읽고 `cycle.Err != nil`이면 구조화 로그에
error 등급으로 남긴다. 분기 구조·반환 조건·주기는 바뀌지 않는다.

**a064가 바꾸지 않는 것**: 반환 조건. 사이클 실패는 여전히 반환값이 아니고, 여전히
알림도 운영 모드 강화도 아니다 (design D1).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (345) | 무한 for | 사이클 실행 | — | 2.1 |
| B2 (346) | `ctx.Err() != nil` | 없음 | `ctx.Err()` — 정상 종료 | 기존 회귀 |
| B3 (350) | `o.clk.Sleep`이 error | 없음 | sleep의 error — 정상 종료 | 기존 회귀 |

새 분기는 B1 안의 `cycle.Err != nil` 하나이고 side effect는 로그 한 줄이다. B2·B3의
조건과 반환은 그대로다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ctx.Err` | 취소 확인 | — | AST |
| `o.ObserveOnce` | 한 사이클 실행 | 실패는 반환하지 않고 `cycle.Err`에 담는다 | AST |
| `o.Interval` | 주기 | 0 이하는 기본값 | AST |
| `o.clk.Sleep` | 주기 대기 | ctx 취소 시 error | AST |
| `o.logErr` (신규 호출) | 사이클 실패 보고 | nil Logger 안전 (`exitloop.go:1464-1468`) | 신규 |

`o.logErr`는 이미 이 파일에 있고 `o.opts.Log == nil`을 먼저 확인한다. Logger가 없는
테스트 배선에서도 새 호출은 no-op이다.

## State mutations and fallbacks

- 이 함수는 옵저버 상태를 직접 바꾸지 않는다. `ObserveOnce`가 바꾼다.
- 새 로그는 fallback이 아니라 **관측**이다. 어떤 판단도 이 로그에 의존하지 않는다.
- 실패 방향: 로그가 남지 않아도 루프는 계속 돈다. 로그가 보호를 좌우해서는 안 된다.

## Safety conclusion

- Safe edit boundary: for 루프 본문에서 `ObserveOnce`의 반환값을 받아 실패를 로그하는
  분기. 반환 조건·주기·취소 처리는 손대지 않는다.
- High-risk impact: **yes** — 함수는 High-risk이지만 이 편집은 *관측만 추가*하며
  판정·주문·손절 경로에 어떤 값도 공급하지 않는다.
- §0.3: 청산 발의는 `ObserveOnce` 안에서 이미 끝난 뒤이므로 새 로그가 손절의 즉시성에
  앞서지 않는다.
- §0.2: 토글이 없다. 알림 설정과 무관하게 동작한다.
