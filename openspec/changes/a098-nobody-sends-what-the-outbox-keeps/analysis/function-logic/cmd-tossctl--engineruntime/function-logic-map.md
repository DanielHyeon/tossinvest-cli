# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (**331-388**) — 편집 전에는 `331-401`의 같은 함수
- AST evidence: `ast.json` — 편집 후 branches **6** / returns **8** / calls 11 / assignments 7
  (편집 전 5 / 7 / 10 / 6)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: **`8c668aac…`**

## ⛔ 이 번들은 §5.2의 표에 **없던 자리**다 — 왜 생겼는지 먼저 적는다

§5.2는 `cmd/tossctl` diff를 **4.4의 운영자 표면**으로만 적고 있었다. 그런데 배달
실행자를 프로덕션에서 **실제로 돌게 하는** 한 줄은 여기 있다 —
`RuntimeOptions`를 만드는 유일한 자리가 이 함수이기 때문이다(`engine.go:366`).

**그래서 4.2가 이 기존 함수를 편집했고, 그 순간 이 산출물이 필요해졌다.**
`check_analysis`가 그것을 잡았다(*"missing evidence for modified function
cmd/tossctl/engine.go:engineRuntime"*) — **계획이 아니라 검사가 먼저 알아챘다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 조립이 끝난 `*engine.Context` | `runEngineRun` | 각 구성자가 자기 오류를 낸다 — 이 함수는 **전부 즉시 반환**한다 |
| `ectx.Journal`·`ectx.Entry` | **둘 다 non-nil** | `engine.New` | **B6** — `AlertDeliverer`가 `ErrRuntimeUnavailable`로 거절 |
| `clk` | non-nil | 호출자 | 배달 주기의 시계이자 감독 폴의 시계 |
| `logger` | nil 허용 | 호출자 | nil이면 정지가 로그로 안 남는다 — 게이트 래치는 그대로 걸린다 |

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — **편집 후** 분기 6 · 이탈 8.

| Branch | 위치 | Condition | Return |
|---|---|---|---|
| B1 | `:333` | `engineFillDetector` 오류 | `:334` |
| B2 | `:342` | `ReconcileDriver` 오류 | `:343` |
| B3 | `:353` | `ExitObserver` 오류 | `:354` |
| B4 | `:358` | `Recovery` 오류 | `:359` |
| B5 | `:362` | `NewRefreshingPairedStrategyEntrySupervisor` 오류 | `:363` |
| **B6 (a098 신설)** | `:374` | **`ectx.AlertDeliverer(clk)` 오류** | `:375` |
| (분기 아님) | `:378`·`:387` | `NewRuntime(...)` 반환 | — |

**여섯 갈래가 전부 같은 모양이다** — `x, err := 만든다; if err != nil { return nil, err }`.
신설 갈래도 그 모양을 따랐고, **다른 선택지가 있었다**: 배달 실행자를 못 만들면
그것만 빼고 나머지로 엔진을 세우는 것. **안 골랐다.**

> **왜 fail-closed 인가.** 배달 실행자가 없는 엔진은 **정확히 a098 이전의 엔진**이다 —
> critical 알림을 원장에 쓰고 아무도 안 보내는 상태. 그것을 조용히 허용하면
> 이 change 가 고친 결함이 **오류 하나를 삼키는 것만으로** 돌아온다.
> 그리고 `AlertDeliverer`가 거절하는 유일한 조건은 *"원장이나 게이트가 없다"*이므로,
> 그 상태에서 세운 엔진은 어차피 매매를 못 한다.

## Calls and live bindings

| Callee | Why | 오류 계약 |
|---|---|---|
| `engineFillDetector` | 체결 감지 | B1 |
| `ectx.ReconcileDriver` | 대사 루프 | B2 |
| `ectx.ExitObserver` | exit 관측 루프 | B3 |
| `ectx.Recovery` | 재기동 복구 시퀀스 | B4 |
| `ectx.NewRefreshingPairedStrategyEntrySupervisor` | `strategy-entry` 외곽 루프 | B5 |
| **`ectx.AlertDeliverer` (a098)** | **밀린 critical 알림의 발송 주체** | **B6** |
| `engine.NewRuntime` | 감독 넷 + **보조 하나** | 조립 검증 |

## State mutations and fallbacks

이 함수는 **조립**이다. 원장에 쓰지 않고 goroutine 을 띄우지도 않는다 —
띄우는 것은 `Runtime.Run`이다.

| Mutation | 무엇 | Fallback |
|---|---|---|
| `RuntimeOptions.Loops` | 감독 루프 **넷** — 이 change 가 **한 자도 안 바꾼다** | — |
| **`RuntimeOptions.Auxiliary` (a098)** | 보조 실행자 **하나** | 없다 — 못 만들면 조립을 거절한다(B6) |

**`Loops` 리터럴은 이 편집에서 안 움직였다.** 그것이 R12 가 지는 성질이고,
`TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop`
(`cmd/tossctl/engine_strategy_entry_dormant_test.go:50`)의 `reflect.DeepEqual`이
**순서까지** 고정한다.

## Safety conclusion

- Safe edit boundary: 구성자 호출 하나 + 갈래 하나 + `Auxiliary` 필드 하나.
  **`Loops` 리터럴과 다른 다섯 갈래는 안 건드린다.**
- High-risk impact: **yes** — 엔진이 무엇을 띄우는지 정하는 자리다.
  방향은 **보수적**이다(fail-closed 하나 추가, 기존 판정 0건 변경).
- 되돌리기: 그 갈래와 `Auxiliary` 한 줄을 지우면 오늘로 돌아간다 —
  **그리고 오늘이 「아무도 안 보낸다」다.** 뮤테이션 L 이 그것을 실제로 냈다.
