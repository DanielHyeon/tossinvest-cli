# Function Logic Map: `TestTheSameNumberOfPlacesIsADifferentMoveInADifferentList`

- Source: `internal/candidate/metrics_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L734–763, 분기 4개)
- Risk scan: `risk-pattern-report.md`

같은 칸 수 이동이 목록 길이에 따라 다른 사건이라는 것(D8). 이 change가 **전제를 실재하는
것으로 바꿨다**.

전에는 150행 목록과 100행 목록을 비교하며 주석에 "KR 패널은 150행을 돌려준다"라고 적었다.
이 시스템의 어떤 패널도 150행을 돌려준 적이 없다. 산술은 옳았고 전제가 허구였다.

지금 비교하는 두 길이는 **실제로 존재하는 둘**이다 — 공식 순위 100행과 WTS 인기 30행 —
그리고 그것을 선언하는 두 소스에서 읽은 값이다. 3칸 이동이 30행에서 10%p, 100행에서 3%p다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `popular` | 30행 목록, 21위 → 18위 | `SourceWTSPopular` | gain 10 |
| `official` | 100행 목록, 21위 → 18위 | `SourceOfficialTradingValue` | gain 3 |
| 두 길이 | 30과 100 — 소스가 선언한 값 | `candidatesrc.Panel` | `panelsize_drift_test.go`가 주석과 소스의 일치를 강제 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!popularMove.Computed() || !officialMove.Computed()` | 없음 | `t.Fatalf` — 둘 다 계산돼야 비교가 의미를 갖는다 | 자체 실행 |
| B2 | `popularMove.PercentileGain != "10"` | 없음 | `t.Errorf` | 자체 실행 |
| B3 | `officialMove.PercentileGain != "3"` | 없음 | `t.Errorf` | 자체 실행 |
| B4 | 두 gain이 같다 | 없음 | `t.Error` — 정규화가 없으면 같아진다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `seriesOf(t, rankObs(...), rankObs(...))` | 두 시리즈 | `t.Fatal` | ast.json calls |
| `RankChange(...)` | 측정 대상 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — fixture의 두 길이와 기대값, 주석의 전제.
- High-risk impact: no (테스트 전용). 정규화 자체는 유지된다 — WTS 30행과 공식 100행이 한 저장소에 섞이므로 필요는 실재하고, 근거로 적혀 있던 예시만 틀렸다.
