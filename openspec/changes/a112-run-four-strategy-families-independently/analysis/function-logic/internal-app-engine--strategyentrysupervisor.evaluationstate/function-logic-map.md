# Function Logic Map: `StrategyEntrySupervisor.evaluationState`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: base — 이 change 는 이 함수를 **고치지 않는다.** 태스크 5.6 이
  runMarket 의 `800-801`(잠긴 시장이 큐에 남은 요청을 건너뛴다)과
  `813-814`(refresh-only worker 가 오류를 삼킨다)를 채우려면, 두 갈래를 여는
  판정이 여기 있다는 것을 열거해야 한다.

## CodeGraph hard evidence

| 관계 | 결과 |
|---|---|
| callers | `runMarket` (`:769`) — 유일하다 (`:783`) |
| callees | `s.mu.RLock`/`RUnlock`, `s.clk.Now`, `Before` |

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `worker` | non-nil | `s.workers[market]` | nil 이면 B1 이 `(false,false)` — 평가하지 않는다 |
| `s.accepting` | 종료 전 true | `Run` 이 열고 `stopAcceptingAndDrain` 이 닫는다 | 닫히면 남은 큐 항목이 전부 B1 로 떨어진다 |
| `worker.latched` | 잠금은 되돌릴 수 없다 | `latchMarket` (`937`) | true 면 B1 — **이것이 `800-801` 을 여는 조건이다** |
| `worker.descriptor.RefreshesAuthority` | 생산 부트스트랩은 true | `NewRefreshingPairedStrategyEntrySupervisor` (`:346`) | true 면 만료를 보지 않고 항상 허용 (B2) |
| `worker.descriptor.AuthorityExpiresAt` | effective worker 는 미래 | `strategyaccount` 권위의 `FreshUntil()` (`:418`) | 지나면 `(false,true)` — 호출자가 만료 잠금을 건다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`924:2`) | 종료 중이거나 nil 이거나 **잠겼거나** 사이클이 없거나 (dormant 이고 갱신자도 아님) | 없음 (RLock 만) | `(false, false)` (`895:3`) | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` |
| B2 (`928:2`) | 권한 갱신 worker | 없음 | `(true, false)` (`898:3`) | `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError` |
| 본문 (`901:2`) | effective worker | 없음 | `(fresh, !fresh)` — 만료 여부가 두 번째 값 | `TestExpiredAuthorityLatchesBeforeEvaluation` |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `s.mu.RLock` | 922:2 | 읽기 잠금 — 이 함수는 아무것도 바꾸지 않는다 |
| `s.mu.RUnlock` | 923:8 | `defer` |
| `Before` | 931:11 | 권한 신선도 비교 |
| `s.clk.Now` | 931:11 | 주입 시계 |

Exact AST return positions: 926:3, 929:3, 932:2.

## State mutations and fallbacks

- 상태 변경 없음. 이 함수는 **판정만** 한다. 그것이 `runMarket` 이 큐에서 항목을
  먼저 꺼낸 뒤에 이 판정을 부를 수 있는 이유이고, 동시에 `800-801` 이 필요한
  이유다 — 항목은 이미 소비되었으므로 거절은 그 항목을 **버리는** 것이다.

## Safety conclusion

- Safe edit boundary: 편집 없음. 인용만.
- High-risk impact: yes — 진입 평가가 도는지 마는지를 이 함수가 정한다.
- **인용해 가는 사실:** 순서가 계약이다. 엔진은 **큐를 먼저 읽고 그다음 거절**한다
  (`812:3` → `783`). 5.3.2 의 `strategyworker.Lane` 은 반대로 **잠금을 먼저 보고
  트리거를 건드리지 않는다.** 두 순서가 다른 것은 실수가 아니라 트리거가 무엇을
  들고 있느냐의 차이다 — 엔진의 트리거는 빈 `struct{}` 라 버려도 잃을 것이 없다.
