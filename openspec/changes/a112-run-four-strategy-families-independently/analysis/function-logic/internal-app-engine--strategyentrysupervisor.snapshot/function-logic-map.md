# Function Logic Map: `StrategyEntrySupervisor.Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: **modified (태스크 8.8.4, 2026-09-05).** 편집은 **한 줄**이다 — 반환하는
  값에 `SwallowedCycleErrors`·`FirstSwallowedFailure` 를 싣는다. 분기·반환 자리·잠금
  자세는 그대로이고, 새 필드는 이미 잡고 있는 `s.mu.RLock` 아래에서 읽힌다.
- Source SHA-256: `c713aa34dee53aa7d276e488efbd6928bc1be598e94128f66fc7f0055de392e0`
- Lines: 727-748

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 이거나 살아 있는 감독자 | 호출자 | nil 이면 영값 + `false` (`729:3`) |
| `market` | `validStrategyMarket` 이 참인 값만 | `strategy_entry_supervisor.go` 의 시장 열거 | 아니면 영값 + `false` (`729:3`) |
| `s.workers[market]` | 등록된 시장이면 non-nil | `NewStrategyEntrySupervisor` 가 만든 표 | nil 이면 영값 + `false` (`735:3`) |
| `worker.swallowedCount` | 0 이상, `math.MaxUint64` 에서 포화 | `recordSwallowedCycleError` 하나 | 포화하며 되감기지 않는다 |
| `worker.firstSwallowed` | 빈 문자열이거나 **첫** 원인 | 같은 함수 | 이후 원인은 덮어쓰지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`728:2`) | `s == nil \|\| !validStrategyMarket(market)` | 없음 | 영값 + `false` (`729:3`) | 기존 감독자 시험 |
| B2 (`734:2`) | `worker == nil` — 등록되지 않은 시장 | 없음 | 영값 + `false` (`735:3`) | 기존 감독자 시험 |
| (분기 없음) | 정상 경로 | 없음 — **읽기 전용** | 스냅샷 + `true` (`737:2`) | `TestARefreshOnlyWorkerCountsTheCycleErrorsItSwallows` (8.8.4) |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `validStrategyMarket` | 728:18 | 시장 열거 밖의 값을 거른다 — 순수 함수, 오류 없음 |
| `s.mu.RLock` | 731:2 | worker 상태를 **공유 읽기**로 본다. 8.8.4 의 두 필드도 이 잠금 아래에서 읽힌다 |
| `s.mu.RUnlock` | 732:8 | `defer` 로 해제 |
| `len` | 745:15 | 큐 깊이 — 순수 |
| `cap` | 745:49 | 큐 용량 — 순수 |

## State mutations and fallbacks

- **없다.** 이 함수는 읽기만 한다 — 그것이 `StrategyWorkerSnapshot` 이 "activation,
  release, mutation 메서드를 갖지 않는다" 는 계약의 실행 쪽이다.
- 8.8.4 가 더한 두 필드도 같은 성질이다: `recordSwallowedCycleError` 가 쓰기를
  전담하고 여기서는 이미 잡고 있는 읽기 잠금 아래에서 복사만 한다. 별도 잠금을
  더하지 않았고, 더했다면 그것이 새 교착 자리가 됐을 것이다.

## Safety conclusion

- Safe edit boundary: 반환 구조체에 **읽기 전용 필드 둘**을 더하는 것까지다. 이
  함수가 쓰기를 갖게 되면 스냅샷이 관측이 아니라 권한이 된다.
- High-risk impact: no — 주문·손절·사이징 어느 경로도 이 값을 읽지 않는다. 소비자는
  운영 화면과 시험뿐이다.
