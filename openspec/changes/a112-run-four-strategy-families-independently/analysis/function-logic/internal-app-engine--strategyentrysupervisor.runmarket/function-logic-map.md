# Function Logic Map: `StrategyEntrySupervisor.runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: **modified (태스크 5.6.2.1, 2026-09-30).** 그 앞 편집은 8.8.4(2026-09-05 — B12 에 기록 호출). 5.6.2.1 은 B12
  본문 안에 새 분기 B13 을 넣었다: 중앙 무결성 오류면 `blockEntryOnCentralIntegrity`(새 함수 — base 에 없어 FLM
  `not-applicable`)가 신규 진입을 닫고 `continue`, 게이트가 없을 때만 `signalCentral` → `return`. 사람 결정 (6): fail-closed 의
  수단은 EntryGate. 판정 순서(refreshOnly 가 중앙 판정보다 앞)는 루프 생존을 위해 그대로다. 편집 전 번들
  `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`. 재번호는 `branch-test-map.md` 머리글.
- Source SHA-256: `da4fa6d1b57217a08a05d0ae57a4f4be1e3c173c35947e06527b02014ae75b23` · 범위 :878–964 · 분기 17
- **아래 표의 줄 번호(`:770` 등)는 5.3.2 작성 당시 좌표다** — 현재 좌표는 `ast.json`(정본)과 `branch-test-map.md`.

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
| B12 (`:813`) | `refreshOnly` — 권한 갱신 전용 worker 의 오류 | `recordSwallowedCycleError` (포화 계수 + 첫 원인 보존) | `continue` — **시장을 잠그지 않는다** | 스냅샷의 `SwallowedCycleErrors`· |
| **B13 (새, 5.6.2.1)** | B12 안: `isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker)` | 게이트가 있으면 `EntryGate.Block(ReasonStrategyCentralIntegrity, 고정 문구)` 뒤 조건 거짓 → `continue` | 게이트가 없을 때만 `signalCentral` → `return` | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` · `TestWithoutAnEntryGateACentralFaultIsNotSwallowed` · 대조 `TestAnOrdinaryRefreshOnlyCycleErrorDoesNotBlockEntry` |
| B14 (옛 B13, `:816`) | (effective worker) 중앙 무결성 오류 | `s.signalCentral` | `return` — 프로세스 전체 fail-closed(생산 0 — 활성화 로트 전 처분은 이월 표) | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety` |
| B15 (옛 B14, `:821`) | 잠금 자체가 실패 | `s.signalCentral` | `return` | 5.6.1 이 채움(`TestTheFourEscalations…`) |
| B16 (옛 B15, `:825`) | 재시작 대기 실패 | 조건부 `signalCentral` | `return` | 5.6.1 이 채움 |
| B17 (옛 B16, `:826`) | 그 실패가 ctx 취소 때문 | 없음 | 조용한 `return` | 4 개 시험 |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다(5.6.2.1 편집 뒤 재생성). 손으로 고른 목록이 아니다.

| Callee expression | Position |
|---|---|
| `ctx.Done` | 880:9 |
| `ctx.Done` | 886:10 |
| `s.mu.RLock` | 890:4 |
| `s.mu.RUnlock` | 892:4 |
| `s.evaluationState` | 893:24 |
| `s.latchMarket` | 895:30 |
| `s.signalCentral` | 897:6 |
| `s.waitMarketRestart` | 900:15 |
| `ctx.Err` | 901:9 |
| `s.signalCentral` | 904:6 |
| `invokeBoundedStrategyCycle` | 912:43 |
| `s.markAbandoned` | 914:5 |
| `s.recordSwallowedCycleError` | 933:5 |
| `isCentralStrategyIntegrity` | 940:8 |
| `s.blockEntryOnCentralIntegrity` | 940:44 |
| `s.signalCentral` | 941:6 |
| `isCentralStrategyIntegrity` | 946:7 |
| `s.signalCentral` | 947:5 |
| `s.latchMarket` | 950:29 |
| `s.signalCentral` | 952:5 |
| `s.waitMarketRestart` | 955:14 |
| `ctx.Err` | 956:8 |
| `s.signalCentral` | 959:5 |

Exact AST return positions: 881:3, 887:4, 898:6, 902:7, 905:6, 917:5, 942:6, 948:5, 953:5, 957:6, 960:5.

## State mutations and fallbacks

- `s.markAbandoned` 는 `worker.abandoned` 를 `s.mu` 아래에서 true 로 만든다 (`:873`).
  한 번 true 면 되돌리는 코드가 없다.
- `s.latchMarket` 은 첫 실패 이유를 보존한다 — 뒤따르는 실패가 덮어쓰지 않는다.
- **fallback 없음.** 큐가 가득 차서 버려진 투입은 producer 쪽에서 typed 결과
  (`StrategyTriggerFull`, `:206`)로만 돌아가고 **버린 수를 세는 계수기가 없다.**
  골든 `queue.overflow` 는 "typed refusal **and bounded drop counter**" 를 요구한다.

## Safety conclusion

- Safe edit boundary: 5.6.2.1 은 B12 안에 B13 하나만 더했다 — 판정 순서 · 다른 열여섯 분기 · 시장 잠금 규칙 불변. 중앙 무결성 오류가 신규 진입을 닫는 것은 보수 방향(진입만 줄임)이고 청산 · 손절 루프는 영향 없음(Run 불반환).
- High-risk impact: yes — 진입 잠금과 중앙 고장 전파가 여기 있다.
- **인용해 가는 계약 셋:** (1) 단일 비행은 플래그가 아니라 소비자 goroutine 하나로
  성립한다, (2) 카덴스 기준점은 사이클 완료가 아니라 투입 시도다, (3) 넘친 투입의
  수를 세는 곳이 없다.
