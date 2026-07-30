# Function Logic Map: `signalsMarketFrom`

- Source: `internal/console/signals.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L592–608, 분기 2개)
- Risk scan: `risk-pattern-report.md`

한 시장 블록을 조립한다. 이 change가 **한 줄**을 더했다:

```
out.Sightings = signalsSightingSourcesFrom(m.Sightings)
```

veto census는 사유별 건수를 준다. 그것으로는 "공식 거래대금 순위가 짧게 온다"와 "WTS 인기
목록이 짧게 온다"를 구분할 수 없는데, 둘은 가서 볼 곳이 다르고 그중 하나는 `seen_late`
분포 전체가 비는 원인이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.Why` | 비어 있으면 정상 | seam | 비어 있지 않으면 나머지를 조립하지 않는다 |
| `m.Sightings` | 소스별 census | `candidate.TallySightingSources` | 비면 블록이 렌더되지 않는다 |
| `now` | 페이지 instant | `Console.now()` | 행의 나이 계산 기준 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.TrimSpace(m.Why) != ""` | `out.Read`를 미측정으로 | early return — **나머지 블록을 조립하지 않는다** | `TestEveryUnmeasuredStateOnTheSignalsScreenCarriesACodeAndASentence` |
| B2 | `for _, verdict := range m.Verdicts` | 행 append | — | `signals_test.go` 전반 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `signalsPanelFrom` / `signalsVetoTallyFrom` / `signalsAccelTallyFrom` / `signalsBandTalliesFrom` | 네 블록 | 순수 | ast.json calls |
| `signalsSightingSourcesFrom(m.Sightings)` **(신규)** | 소스별 census | 순수 | ast.json calls |
| `signalsRowFrom(verdict, now)` / `sortSignalsRows` | 행과 정렬 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — view 값 하나를 만든다.
- fallback 없음. 읽지 못한 시장은 사유를 실은 미측정이다.

## Safety conclusion

- Safe edit boundary: 본문 1줄 + view 구조체에 필드 1개. 기존 두 분기와 네 블록 무변경.
- High-risk impact: no (렌더 조립).
