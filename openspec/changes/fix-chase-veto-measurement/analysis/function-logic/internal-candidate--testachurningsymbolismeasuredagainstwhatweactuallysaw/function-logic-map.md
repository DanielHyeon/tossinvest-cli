# Function Logic Map: `TestAChurningSymbolIsMeasuredAgainstWhatWeActuallySaw`

- Source: `internal/candidate/metrics_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L773–794, 분기 3개)
- Risk scan: `risk-pattern-report.md`

목록에서 빠졌다 돌아온 종목이 **실제로 본 것**에 대해 측정된다는 것. 이 change가 목록 길이를
150에서 100으로 고치고(기대 gain `1.333333333333` → `2`), 단언을 `.Yes()`로 바꿨다.

`!got.NewlyListed`가 아니라 `!got.NewlyListed.Yes()`인 이유는 앞의 형태가 **아무도 재지
않은 값**에 대해서도 통과했기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시리즈 | 99/100 → 97/100, 두 번째가 신규 진입 | `rankObs` | — |
| 기대 gain | `"2"` — 100행에서 2칸 | 이 테스트 | `t.Errorf` |
| 기대 사실 | `NewlyListed.Yes()` | 이 테스트 | `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!got.Computed()` | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `!got.NewlyListed.Yes()` | 없음 | `t.Errorf` — 시리즈에 구멍이 있다는 기록 | 자체 실행 |
| B3 | `got.PercentileGain != "2"` | 없음 | `t.Errorf` — '직전 최하위'가 credited 했을 값이 아니다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `seriesOf(...)` / `RankChange(...)` | 시리즈와 측정 | `t.Fatal` | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 숫자 3곳 + 단언 형태 1곳.
- High-risk impact: no (테스트 전용).
