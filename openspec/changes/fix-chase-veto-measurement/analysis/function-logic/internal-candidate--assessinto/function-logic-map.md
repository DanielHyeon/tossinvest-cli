# Function Logic Map: `assessInto`

- Source: `internal/candidate/watch.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L669–689, 분기 2개)
- Risk scan: `risk-pattern-report.md`

한 turn의 결과에 읽기 전용 assessment를 붙인다. 이 change가 **한 줄**을 더했다:

```
res.Sightings = TallySightingSources(verdicts)
```

`VetoTally`는 사유별 건수를 준다. 그것으로는 "공식 거래대금 순위가 짧게 온다"와 "WTS 인기
목록이 짧게 온다"를 구분할 수 없는데, 둘은 가서 볼 곳이 다르고 그중 하나는 `seen_late`
분포 전체가 비는 원인이다. 이 한 줄이 `CycleResult`에 그 census를 싣는다.

reducer는 `internal/candidate`에 있고 두 표면(`tossctl candidate scan`, `/signals`)이
**같은 것**을 쓴다 — 이 저장소가 네 번째 그림자 밴드에서 한 번 지불한 규칙이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.SkipAssessment` | 조용한 turn 여부 | `Cycle` | true면 이전 결과를 그대로 둔다 |
| `opts.Thresholds` | 적용할 임계 | `candidateVetoThresholds()` | 부재는 `THRESHOLD_ABSENT` |
| `at` | assessment 기준 instant | `Cycle` | — |
| `verdicts` | `Assess`의 출력 | `Assess` | 이 함수의 세 reducer 입력 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `opts.SkipAssessment` | 없음 | `res, nil` — 지난 turn이 찾은 것을 그대로 들고 있다 | `TestATurnWithNothingDueIsNotAMarketFailure` |
| B2 | `Assess` 오류 | 없음 | `res, err` | **커버 없음** — `Assess`를 실패시키는 fixture가 없다 |

조용한 경로와 보통 경로가 **같은 함수**를 지나는 것이 의도다 — 아무것도 due하지 않은
turn도 지난 turn이 찾은 것을 전부 들고 있어야 하고, 거기서 0개의 verdict를 받은 호출자는
빈 시장을 렌더한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Assess(ctx, store, AssessOptions{...})` | 판정 | 오류는 그대로 상승 | ast.json calls |
| `TallyVerdicts(verdicts)` | veto·crossing·band 세 tally | 순수 | ast.json calls |
| `TallySightingSources(verdicts)` **(신규)** | 소스별 최초 관측 census | 순수 | ast.json calls |

## State mutations and fallbacks

- `res.Verdicts`, `res.Vetoes`, `res.Crossings`, `res.Bands`, **`res.Sightings`**(신규) — 값 복사에 대입.
- 저장소에 쓰지 않는다. 이름 그대로 읽기 전용 assessment다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 본문 1줄 + `CycleResult`에 필드 1개. 기존 세 tally와 두 분기 무변경.
- High-risk impact: no (읽기 전용 집계). 조용한 경로와 보통 경로가 갈라지면 `/signals`가 turn마다 깜빡인다 — 기존 위험이며 이 change가 건드리지 않았다.
