# Function Logic Map: `StrategyEntrySupervisor.runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: **modified (태스크 8.8.4, 2026-09-05).** 그전까지 base 였고, 처음에는
  태스크 5.3.2 가 이 루프를 영수증으로 인용하려고 만든 번들이었다. 8.8.4 가 B12 에
  기록 호출 하나를 더했다 — 버리기 **전에** 센다. 삼킴 자세(`continue`, 잠그지 않음,
  갈래 순서)는 바꾸지 않았고 그것을 못 박는 기존 시험 둘이 그대로 통과한다.
- Source SHA-256: `c2a825fceac8692f865d89ae61b736d9a43ad10e0b5dec4968cfac304c93555a`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `start` | 닫히는 배리어 채널 | `Run` 이 두 시장 자식을 함께 푼다 (`:695`) | 닫히기 전 ctx 취소면 사이클 0 회로 끝난다 (`:772`) |
| `worker.queue` | `chan struct{}`, 용량 `depth` | 생산은 `DefaultStrategyQueueDepth` = 1 (`:510`–`:512`, `:569`) | 가득 차면 producer 가 `StrategyTriggerFull` 을 받고 **버린 수를 세지 않는다** |
| `worker.effective`, `worker.latched` | `s.mu` 아래 | `evaluationState` (`:852`) | 잠긴 worker 는 `allowed=false` 로 사이클을 건너뛴다 |
| `s.cycleLimit` | 양의 유한 시간 | 생산 `MaximumStrategyCycleLimit` = 30s | `invokeBoundedStrategyCycle` 의 마감 시한 |

**이 루프가 단일 비행인 이유는 플래그가 아니라 구조다.** 시장마다 이 함수를 도는
goroutine 이 **하나**이고, 사이클은 `<-worker.queue` 를 다시 읽기 전에 반드시
반환한다(`:779` → `:803` → 루프 끝 `:832`). 즉 "동시에 두 사이클"은 이 코드에서
표현할 수 없다. 이것은 검사로 지켜지는 것이 아니라 배선으로 지켜지는 성질이라,
드라이버가 둘이 되는 순간 조용히 사라진다.

**카덴스도 이 함수에 없다.** 주기는 `runStrategyPoller` (`:719`)가 갖고 있고,
그 goroutine 은 enqueue 를 시도한 **직후** `clk.Sleep(ctx, interval)` 한다 —
사이클이 끝난 뒤가 아니다. 그래서 주기의 기준점은 사이클 완료가 아니라 투입 시도다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`:770`) | 배리어 전 ctx 취소 | 없음 | `return` (`:772`) | 배리어 경합 시험 |
| B3 (`:776`) | ctx 취소 vs 큐 도착 | 없음 | `return` (`:778`) 또는 사이클 진행 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` |
| B4 (`:785`) | 권한 만료 | `s.latchMarket` — 시장 잠금 | 잠근 뒤 backoff 대기, `continue` | `TestExpiredAuthorityLatchesBeforeEvaluation` |
| B5 (`:787`) | 만료 잠금이 실패 | `s.signalCentral` | `return` (중앙 고장) | **없음 — 측정으로 확인** |
| B6 (`:791`) | 만료 뒤 재시작 대기 실패 | 조건부 `signalCentral` | `return` | `TestExpiredAuthorityLatchesBeforeEvaluation` |
| B7 (`:792`) | 그 실패가 ctx 취소 때문 | 없음 | 조용한 `return` (`:793`) | `TestExpiredAuthorityLatchesBeforeEvaluation` |
| B8 (`:800`) | `!allowed` — 잠겼거나 꺼졌거나 cycle 이 nil | 없음 | `continue` — 사이클 없음 | **없음 — 측정으로 확인** |
| B9 (`:804`) | `abandoned` | `s.markAbandoned` | 계속 | 4 개 시험 |
| B10 (`:807`) | `cancelled` | 없음 | `return` | 3 개 시험 |
| B11 (`:810`) | `err == nil` | 없음 | `continue` | 3 개 시험 |
| B12 (`:813`) | `refreshOnly` — 권한 갱신 전용 worker 의 오류 | `recordSwallowedCycleError` (포화 계수 + 첫 원인 보존) | `continue` — **잠그지 않는다** | 스냅샷의 `SwallowedCycleErrors`·`FirstSwallowedFailure` (8.8.4) |
| B13 (`:816`) | 중앙 무결성 오류 | `s.signalCentral` | `return` — 모든 신규 진입 정지 | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety` |
| B14 (`:821`) | 잠금 자체가 실패 | `s.signalCentral` | `return` | **없음 — 측정으로 확인** |
| B15 (`:825`) | 재시작 대기 실패 | 조건부 `signalCentral` | `return` | **없음 — 측정으로 확인(`:829` 블록)** |
| B16 (`:826`) | 그 실패가 ctx 취소 때문 | 없음 | 조용한 `return` | 4 개 시험 |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다. 손으로 고른 목록이 아니다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `ctx.Done` | 856:9 | 배리어 전 취소 |
| `ctx.Done` | 862:10 | 루프마다 취소 확인 |
| `s.mu.RLock` | 866:4 | `refreshOnly` 를 읽기 위한 공유 잠금 |
| `s.mu.RUnlock` | 868:4 | 같은 잠금 해제 |
| `s.evaluationState` | 869:24 | 이 사이클을 돌려도 되는지 + 권한 만료 여부 |
| `s.latchMarket` | 871:30 | 권한 만료로 시장 진입 잠금 |
| `s.signalCentral` | 873:6 | 잠금 자체가 실패하면 중앙 고장 |
| `s.waitMarketRestart` | 876:15 | backoff 만큼 대기 |
| `ctx.Err` | 877:9 | 그 대기 실패가 취소 때문인지 |
| `s.signalCentral` | 880:6 | 아니면 중앙 고장 |
| `invokeBoundedStrategyCycle` | 888:43 | **마감 시한 아래 사이클 한 번.** 이 함수의 유일한 호출자다(CodeGraph) |
| `s.markAbandoned` | 890:5 | 마감 시한을 넘긴 사이클을 버려진 것으로 표시 |
| `s.recordSwallowedCycleError` | 909:5 | **버리기 전에 센다** (8.8.4). `continue` 도 갈래 순서도 바꾸지 않는다 — 더한 것은 기록뿐이다 |
| `isCentralStrategyIntegrity` | 912:7 | 원장·게이트웨이·펜스·소유자 무결성 오류인지 |
| `s.signalCentral` | 913:5 | 그러면 모든 신규 진입을 멈춘다 |
| `s.latchMarket` | 916:29 | 보통 오류로 시장 진입 잠금 |
| `s.signalCentral` | 918:5 | 잠금 자체가 실패하면 중앙 고장 |
| `s.waitMarketRestart` | 921:14 | backoff 만큼 대기 |
| `ctx.Err` | 922:8 | 그 대기 실패가 취소 때문인지 |
| `s.signalCentral` | 925:5 | 아니면 중앙 고장 |

Exact AST return positions: 857:3, 863:4, 874:6, 878:7, 881:6, 893:5, 914:5, 919:5, 923:6, 926:5.


## State mutations and fallbacks

- `s.markAbandoned` 는 `worker.abandoned` 를 `s.mu` 아래에서 true 로 만든다 (`:873`).
  한 번 true 면 되돌리는 코드가 없다.
- `s.latchMarket` 은 첫 실패 이유를 보존한다 — 뒤따르는 실패가 덮어쓰지 않는다.
- **fallback 없음.** 큐가 가득 차서 버려진 투입은 producer 쪽에서 typed 결과
  (`StrategyTriggerFull`, `:206`)로만 돌아가고 **버린 수를 세는 계수기가 없다.**
  골든 `queue.overflow` 는 "typed refusal **and bounded drop counter**" 를 요구한다.

## Safety conclusion

- Safe edit boundary: 이 change 는 이 함수를 편집하지 않는다. 인용만 한다.
- High-risk impact: yes — 진입 잠금과 중앙 고장 전파가 여기 있다.
- **인용해 가는 계약 셋:** (1) 단일 비행은 플래그가 아니라 소비자 goroutine 하나로
  성립한다, (2) 카덴스 기준점은 사이클 완료가 아니라 투입 시도다, (3) 넘친 투입의
  수를 세는 곳이 없다.
