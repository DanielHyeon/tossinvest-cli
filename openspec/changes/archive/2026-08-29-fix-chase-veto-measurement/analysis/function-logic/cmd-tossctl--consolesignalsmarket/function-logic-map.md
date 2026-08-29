# Function Logic Map: `consoleSignalsMarket`

- Source: `cmd/tossctl/console.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L884–917, 분기 1개)
- Risk scan: `risk-pattern-report.md`

`/signals` seam이 한 시장을 판정한다. 이 change가 **두 줄**을 바꿨다.

- `Thresholds: candidateVetoThresholds()` — 여기 있던 `VetoThresholds{…}` 리터럴을 지웠다.
  같은 리터럴이 `candidate.go`에도 있었고, 한쪽만 편집되면 `/signals`와
  `tossctl candidate scan`이 같은 저장소·같은 시각에 대해 다른 답을 낸다.
- `out.Sightings = candidate.TallySightingSources(verdicts)` — 스캔 출력과 **같은 reducer**.

`/signals`는 스캔 기록을 볼 수 없으므로 소스별 요청·도착(`ScanResult.Readings`)은 렌더할 수
없고 census만 렌더한다. 그 한계는 issues.md I9에 적혀 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `store` | 열린 discovery store | `consoleSignals.open` | — |
| `market` | 계약 시장 | `consoleSignalsMarkets` | — |
| `at` | **모든 시장에 하나** | `store.Now()` | 두 시장이 두 instant로 읽히면 같은 질문에 다른 답이 나온다 |
| 임계 | `candidateVetoThresholds()` — 스캔과 같은 함수 | `vetothresholds.go` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Assess` 오류 | `out.Why` 설정 | 부분 결과 — 시장이 **누락되지 않고** 사유와 함께 남는다 | **커버 없음** — `Assess`를 실패시키는 seam fixture가 없다 |

읽지 못한 시장이 화면에서 **사라지지 않는** 것이 이 함수의 기존 결정이고 이 change가
건드리지 않았다 — 없는 시장과 아무것도 없는 시장은 구분할 수 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidate.Assess(ctx, store, AssessOptions{...})` | 판정 | 오류는 `Why` | ast.json calls |
| `candidateVetoThresholds()` | 임계 단일 출처 | 순수 | ast.json calls |
| `candidate.TallyVerdicts(verdicts)` | 세 tally — 손으로 조립하지 않는다 | 순수 | ast.json calls |
| `candidate.TallySightingSources(verdicts)` **(신규)** | 소스별 census — 스캔과 같은 reducer | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 값 하나를 만든다. 저장소에 쓰지 않는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 임계 리터럴 → 함수 호출, `Sightings` 대입 1줄.
- High-risk impact: **yes 인접** — chase veto의 입력을 만든다. 이 change는 값을 바꾸지 않고 출처를 하나로 만들었으며, 그 방향은 두 화면이 갈라지는 것을 막는 쪽이다.
