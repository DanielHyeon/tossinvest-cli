# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (356-443)
- AST evidence: `ast.json` — AST 기준 branches **6** / returns 7 / calls 12
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `8ad1cc88b9e0a0181c0a1c50a1fbcbdad4a6f85a6786e316c527ba5565274110`
- 작성 사유: a102 §3(D5·D5b) — `Recover` 클로저가 편집 전 `recovery.Run`의 Report를 버렸다
  (1판 `:402`). D5는 그 클로저를 `recoverThenReady`로 교체하고, D5b는 버려지던 Report를 소비한다.
  **기존 함수의 내부를 편집하므로 편집 전에 만든다.** 루프 집합 조립이므로 High-risk다.

> **두 판을 병기한다.** 1판(편집 전)은 `:347-430` · 분기 6 · 이탈 **8** · 호출 11 ·
> SHA `f13e36b35e08…` · 블록 14개 중 10개 실행이었고, 2판(GREEN 후, 이 문서)은
> `:356-443` · 분기 **6**(그대로) · 이탈 **7** · 호출 **12** · SHA `8ad1cc88b9e0…` ·
> 블록 **13개 중 10개 실행**이다.
>
> **이탈이 8 → 7로 줄었다.** 사라진 것은 `Recover` 클로저 안의 `return rerr`(1판 `:403`)다 —
> 그 클로저가 `recoverThenReady(...)` 호출로 바뀌면서 **판정이 이 함수 밖으로 나갔다.**
> 1판이 측정한 그 클로저의 count=0이 이 형태를 정한 근거였고, 지금 그 자리는 100.0%로
> 측정되는 `recoverThenReady`가 진다.

## 이 함수가 하는 일

세 안전 루프(reconcile · exit · filldetect) + 비활성 전략 진입 루프를 만들고, 기존
all-or-nothing supervisor(`engine.NewRuntime`)에 넘긴다. `Recover`·`Loops`·`Auxiliary`를
채우는 것이 전부이고, **판정은 없다** — 6개 분기가 전부 "만들다 실패하면 즉시 return"이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | `runEngineRun` | 조립에는 쓰이지 않는다(호출자에 전달만) |
| `ectx` | non-nil | `engineAssemble` | 각 `ectx.*` 생성자가 오류를 돌려준다 |
| `clk` | non-nil | `clock.System()` | — |
| `logger` | **nil 가능** | `runEngineRun`이 만든 것 / 테스트는 nil | `obs.Logger`의 메서드가 nil 수신자를 견딘다(`log.go:191`) |
| `recovery` (`:383`) | non-nil | `ectx.Recovery(reconcile.Options{Clock: clk})` | B4가 즉시 return |

> **관통 불변식**: 이 함수는 **아무것도 실행하지 않는다.** 조립뿐이고, 실행은
> `engine.Runtime`이 한다(`internal/app/engine/runtime.go:289-294`가 `opts.Recover`를
> 루프 시작 **전에** 부른다). a102는 그 규율을 지킨다 — `recoverThenReady`도 클로저를
> 돌려줄 뿐 아무것도 실행하지 않는다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:359` | `engineFillDetector` 오류 | — | `:360` |
| B2 | `:368` | `ectx.ReconcileDriver` 오류 | — | `:369` |
| B3 | `:379` | `ectx.ExitObserver` 오류 | — | `:380` |
| B4 | `:384` | `ectx.Recovery` 오류 | — | `:385` |
| B5 | `:388` | `NewRefreshingPairedStrategyEntrySupervisor` 오류 | — | `:389` |
| B6 | `:400` | `ectx.AlertDeliverer` 오류 | — | `:401` |

정상 이탈: `:404` `return engine.NewRuntime(...)`. 1판의 여덟 번째 return(`Recover` 클로저
안의 `:403` `return rerr`)은 **없어졌다** — 그 클로저가 `recoverThenReady(...)` 호출로
바뀌었기 때문이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFillDetector` `:357` | 체결 감지기 | 조립 시점 검증 | ast.json |
| `ectx.ReconcileDriver` `:363` | 대사 루프 | — | ast.json |
| `ectx.ExitObserver` `:372` | exit 관측 루프 | — | ast.json |
| `ectx.Recovery` `:383` | **재시작 복구** | `reconcile.Options{Clock: clk}` | ast.json |
| `ectx.AlertDeliverer` `:399` | 알림 outbox 배출 (Auxiliary) | 전달 실패가 루프를 세우지 않는다 | ast.json |
| `engine.NewRuntime` `:404` | supervisor 조립 | 이 함수의 최종 return | ast.json |
| **`recoverThenReady(recovery.Run, ready, engineRecoveryObserver(logger))` `:417`** | **재시작 복구 배선** | `recovery.Run`을 **부르지 않는다** — 감싼 클로저를 만들 뿐이다 | ast.json |

live binding — 1판 `:402`의 `_, rerr := recovery.Run(ctx)`가 **A1 F1이 지적한 자리**였다.
§1이 최대 5분의 429 대기를 만들었는데 그 대기의 관측치를 받을 소비자가 이 프로세스에 없었다.
2판에서 그 소비자는 `engineRecoveryObserver`이고, 전수는 이제 두 곳이다:
`rg -n 'RateLimitWaits|RateLimitWaited' --glob '!*_test.go'` →
`internal/reconcile/{recovery,ratelimit}.go` + **`cmd/tossctl/engineready.go`**.

## State mutations and fallbacks

- 프로세스 밖 상태를 **바꾸지 않는다.** 조립만 한다.
- fallback이 없다 — 6개 분기 전부 즉시 return이다. **선택 seam이라는 개념이 여기엔 없다.**
- `logger`가 nil이어도 조립은 계속된다(테스트가 그렇게 부른다). a102의 obs 한 줄도 같은
  규율을 따라야 한다: **nil logger가 조립을 깨면 안 된다.**

## Safety conclusion

- Safe edit boundary: **`Recover:` 필드 하나**(1판 `:401-404` → 2판 `:417` 한 줄)와
  **시그니처의 인자 하나**(`ready func()`). 6개 분기의 조건·순서·오류 문구는 불변이고,
  `Loops`·`Auxiliary`의 내용도 불변이다. **실제로 그 둘이었다.**
- 시그니처 변경이 필요한 이유: ready 신호를 만드는 핸들은 `runEngineRun`(6단계)에 있고,
  `Recover` 클로저는 여기(7단계)에서 만들어진다. `engine.Runtime`은 조립 후 `Recover`를
  갈아끼울 수단을 주지 않고, **`internal/app/engine`은 무변경**이 design의 못이다.
  따라서 seam은 인자로 들어온다.
- High-risk impact: **yes** — 손절을 두는 루프 집합의 조립이다. 방향은 보수적이다:
  `recoverThenReady`는 성공에만 `ready()`를 부르므로, 편집이 잘못돼도 최악은
  "ready 신호가 안 뜬다" = **오늘의 동작**(콘솔이 안 기다린다)이다.
- 물려받은 공백: B1·B5·B6은 편집 전·후 모두 count=0이다(`branch-test-map.md`).
  **`Recover` 클로저의 공백은 없어졌다** — 클로저가 사라지고 그 자리에 100.0%로 측정되는
  `recoverThenReady`·`engineRecoveryObserver` 호출이 남았기 때문이다. 미측정 자리를
  메운 것이 아니라 **판정을 미측정 자리에서 빼낸** 것이다.
