# Function Logic Map: `invokeBoundedStrategyCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: base — 이 change 는 이 함수를 **고치지 않는다.** 태스크 5.3.2 가
  레인-로컬 마감 시한을 새로 만들면서 이 함수의 분기를 영수증으로 인용하기 때문에
  만든 번들이다. 손으로 읽은 인용을 근거로 쓰지 않기 위한 AST 열거다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | non-nil, 취소 가능 | `runMarket` 의 child context | 취소되면 `ctx.Err()` 를 cancelled=true, abandoned=true 로 돌려준다 (`:890`) |
| `clk` | non-nil `clock.Clock` | `s.clk` — 저장소의 단일 주입 시계 | 감시견은 `clk.Sleep(watchdogCtx, limit)` 하나뿐이라 시계가 없으면 마감 시한이 없다 |
| `limit` | 양의 유한 시간 | `s.cycleLimit` — 생산은 `MaximumStrategyCycleLimit` = 30s (`:318`, `:352`) | 0 이하면 `clk.Sleep` 가 즉시 반환해 모든 사이클이 마감 초과가 된다 |
| `cycle` | non-nil `StrategyCycle` | `worker.descriptor.Cycle` | nil 은 이 함수 앞 `evaluationState` 가 이미 막는다 (`:857`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-a (`:889`) | `<-ctx.Done()` — 상위가 먼저 취소됨 | 없음. 사이클 goroutine 은 **계속 돈다** | `ctx.Err(), false, true, true` (`:890`) | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` |
| B1-b (`:891`) | `outcome := <-result` — 사이클이 먼저 끝남 | 없음 | `outcome.err, outcome.abnormal, false, false` (`:892`) | `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` |
| B1-c (`:893`) | `<-deadline` — 감시견이 먼저 깨어남 | 없음 | 아래 B2 로 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` |
| B2 (`:894`) | 마감 시한이 울렸는데 `ctx.Err() != nil` | 없음 | `ctx.Err(), false, true, true` (`:895`) | `TestTheWatchdogRechecksCancellationInsteadOfTrustingItsOwnTimer` (5.6) |
| B2-else (`:897`) | 마감 시한이 울렸고 ctx 는 살아 있음 | 없음 | `ErrStrategyCycleDeadline, true, false, true` (`:897`) | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다. 손으로 고른 목록이 아니다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `make` | 1064:12 | 결과 채널. 버퍼 1 이라 늦게 끝난 사이클이 goroutine 을 붙잡지 않는다 |
| `(unnamed)` | 1065:5 | 사이클 goroutine. **이 goroutine 은 마감 시한에 취소되지 않는다** |
| `invokeStrategyCycle` | 1066:13 | panic 을 `abnormal=true` 인 error 로 바꾼다 |
| `context.WithCancel` | 1068:33 | 감시견 전용 취소 문맥 |
| `cancelWatchdog` | 1069:8 | `defer` — 감시견 goroutine 누수를 막는다 |
| `make` | 1070:14 | 감시견 채널. 버퍼 1 |
| `(unnamed)` | 1071:5 | 감시견 goroutine |
| `clk.Sleep` | 1071:26 | **마감 시한 그 자체.** `time.After` 가 아니라 주입 시계(`internal/clock`)다 |
| `ctx.Done` | 1073:9 | 상위 취소 관측 |
| `ctx.Err` | 1074:10 | 취소 사유를 그대로 돌려준다 |
| `ctx.Err` | 1078:6 | 마감 시한이 울린 뒤 상위가 이미 취소되었는지 |
| `ctx.Err` | 1079:11 | 그 경우 마감 시한이 아니라 취소를 돌려준다 |

Exact AST return positions: 1074:3, 1076:3, 1079:4, 1081:3.


## State mutations and fallbacks

- 이 함수는 supervisor 상태를 하나도 바꾸지 않는다. `ast.json` 의 `assignments`
  넷은 전부 지역 변수다. 상태 변경(`markAbandoned`, `latchMarket`)은 호출자가 한다.
- **fallback 없음.** 마감 시한을 넘긴 사이클 goroutine 은 취소되지 않고 버려진다
  (`abandoned`). 즉 "abandon" 은 "일을 멈춘다"가 아니라 "기다리기를 멈추고
  버려진 것으로 표시한다"이다. 늦게 도착한 결과는 버퍼 1 인 `result` 채널에
  들어가고 아무도 읽지 않는다.

## Safety conclusion

- Safe edit boundary: 이 change 는 이 함수를 편집하지 않는다. 인용만 한다.
- High-risk impact: yes — 마감 시한은 진입 잠금으로 이어진다. 다만 이 번들의
  용도는 인용이고, 편집은 없다.
- **인용해 가는 계약 하나:** 마감 시한 초과는 `abnormal=true` 로 돌아온다(`:897`).
  설계 문서의 고장표(`design.md:198`)는 deadline 을 "보통 오류 — 세고 다시 시도"
  줄에 두었으므로 **두 정본이 갈린다.** 생산 임계값이 1 이라 오늘은 두 해석의
  결과가 같지만, 임계값이 1 보다 커지는 순간 갈라진다.

> **5.6.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일의 다른 함수 편집으로 +43줄 이동 · 파일 해시만 바뀜). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 편집으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1 리뷰 수리(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 의 몸통 이동으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 6.3 (c) — 같은 파일 재검증 판정 이동으로 줄만 밀림(이 번들의 해시가 HEAD 파일과 같았음 — 앞 로트들과 같은 동기화)

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)
