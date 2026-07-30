# Function Logic Map: `TestANewlyListedSymbolDoesNotClimbFromLastPlace`

- Source: `internal/candidate/metrics_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L674–718, 분기 6개)
- Risk scan: `risk-pattern-report.md`

신규 진입 종목이 "직전에 최하위였다"로 채워져 최대 상승을 지어내지 않는다는 것. 이 change가
**두 가지**를 고쳤다.

1. 목록 길이 `150 → 100`, 경계 사례 순위 `148 → 98`. **150행 패널은 존재한 적이 없다** —
   공식 순위는 KR·US 모두 100행, WTS 인기는 KR 30행이다(design D9). 산술은 옳았고 전제가
   허구였는데, 허구 쪽이 더 나쁘다: 다음 사람이 인용하는 문장이기 때문이다.
2. `!got.NewlyListed` → `!got.NewlyListed.Yes()`. **이것이 사실의 형태 변화 전부다** —
   `!got.NewlyListed`는 아무도 재지 않은 값에 대해서도 참이었고, 그 둘을 가를 세 번째 답이
   없었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 목록 길이 | 100 — 실제로 존재하는 길이 | `candidatesrc.Panel`의 리터럴 | `panelsize_drift_test.go`가 주석·소스 일치를 강제 |
| `tc.rank` | 1(상위)과 98(경계, 매 스캔 churn) | 이 테스트 | — |
| 기대 | `Computed()==false`, 사유 `NO_PRIOR_RANK`, gain 부재, `NewlyListed.Yes()` | 이 테스트 | `t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, tc := range {...}` — 두 진입 위치 | 없음 | — | 자체 실행 |
| B2 | `got.Computed()` | 없음 | `t.Fatalf` — 지어낸 gain이 있으면 즉시 실패 | 자체 실행 |
| B3 | `got.Reason != NotComputedNoPriorRank` | 없음 | `t.Errorf` | 자체 실행 |
| B4 | `got.PercentileGain != ""` | 없음 | `t.Errorf` — 부재는 빈 문자열이지 0이 아니다 | 자체 실행 |
| B5 | `!got.NewlyListed.Yes()` | 없음 | `t.Errorf` — **측정된 yes**여야지 not-a-no가 아니다 | 자체 실행 |
| B6 | `got.Rank != tc.rank || got.RankTotal != 100` | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `seriesOf(t, rankObs(...))` | 한 관측짜리 시리즈 | `t.Fatal`로 실패 | ast.json calls |
| `RankChange(series, at, DefaultAccelerationWindow)` | 측정 대상 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 숫자 정정 2곳 + 단언 형태 1곳.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk 인접 — 지어낸 최대 상승은 매 스캔 반복되고, 가속의 입력이 된다.
